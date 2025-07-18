package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// ReactionHandler handles like/dislike reactions
type ReactionHandler struct {
	Repo             *repository.ReactionRepository
	PostRepo         *repository.PostRepository
	CommentRepo      *repository.CommentRepository
	NotificationRepo *repository.NotificationRepository
}

func NewReactionHandler(repo *repository.ReactionRepository, postRepo *repository.PostRepository, commentRepo *repository.CommentRepository, notifRepo *repository.NotificationRepository) *ReactionHandler {
	return &ReactionHandler{Repo: repo, PostRepo: postRepo, CommentRepo: commentRepo, NotificationRepo: notifRepo}
}

// React toggles a reaction on a post or comment for the authenticated user
func (h *ReactionHandler) CreateReact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		TargetID     string `json:"target_id"`
		TargetType   string `json:"target_type"`
		ReactionType int    `json:"reaction_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetID == "" || (req.TargetType != "post" && req.TargetType != "comment") || req.ReactionType == 0 {
		utils.ErrorResponse(w, "Invalid request", http.StatusBadRequest)
		return
	}

	prev, newType, toggleErr := h.Repo.ToggleReaction(user.ID, req.TargetType, req.TargetID, req.ReactionType)
	if toggleErr != nil {
		utils.ErrorResponse(w, "Failed to react", http.StatusInternalServerError)
		return
	}

	var ownerID string
	var postID string
	if req.TargetType == "post" {
		post, _ := h.PostRepo.GetByID(req.TargetID)
		if post != nil {
			ownerID = post.UserID
			postID = post.ID
		}
	} else {
		comment, _ := h.CommentRepo.GetByID(req.TargetID)
		if comment != nil {
			ownerID = comment.UserID
			postID = comment.PostID
		}
	}

	if ownerID != "" && ownerID != user.ID {
		var actions []string
		target := "post"
		if req.TargetType == "comment" {
			target = "comment"
		}
		if newType == 0 {
			if prev == 1 {
				actions = append(actions, "unlike_"+target)
			} else if prev == 2 {
				actions = append(actions, "undislike_"+target)
			}
		} else {
			if prev == 1 && newType == 2 {
				actions = append(actions, "unlike_"+target)
			} else if prev == 2 && newType == 1 {
				actions = append(actions, "undislike_"+target)
			}
			if newType == 1 {
				actions = append(actions, "like_"+target)
			} else if newType == 2 {
				actions = append(actions, "dislike_"+target)
			}
		}
		for _, act := range actions {
			n := models.Notification{
				UserID:  ownerID,
				ActorID: user.ID,
				Action:  act,
				IsRead:  false,
			}
			if req.TargetType == "post" {
				n.PostID = &postID
			} else {
				n.PostID = &postID
				n.CommentID = &req.TargetID
			}
			h.NotificationRepo.Create(n)
		}
	}

	var reactions []models.ReactionWithUser
	var err error
	if req.TargetType == "post" {
		reactions, err = h.Repo.GetReactionsByPostWithUser(req.TargetID)
	} else {
		reactions, err = h.Repo.GetReactionsByCommentWithUser(req.TargetID)
	}
	if err != nil {
		utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, reactions, http.StatusOK)
}
