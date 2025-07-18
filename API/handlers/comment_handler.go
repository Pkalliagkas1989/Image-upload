package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

// CommentHandler handles comment related endpoints
type CommentHandler struct {
	CommentRepo      *repository.CommentRepository
	PostRepo         *repository.PostRepository
	NotificationRepo *repository.NotificationRepository
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(repo *repository.CommentRepository, postRepo *repository.PostRepository, notifRepo *repository.NotificationRepository) *CommentHandler {
	return &CommentHandler{CommentRepo: repo, PostRepo: postRepo, NotificationRepo: notifRepo}
}

// CreateComment creates a new comment on a post for the authenticated user
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PostID == "" || req.Content == "" {
		utils.ErrorResponse(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostID:  req.PostID,
		UserID:  user.ID,
		Content: req.Content,
	}

	created, err := h.CommentRepo.Create(comment)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	post, _ := h.PostRepo.GetByID(req.PostID)
	if post != nil && post.UserID != user.ID {
		n := models.Notification{
			UserID:    post.UserID,
			ActorID:   user.ID,
			PostID:    &post.ID,
			CommentID: &created.ID,
			Action:    "comment",
			IsRead:    false,
		}
		h.NotificationRepo.Create(n)
	}

	utils.JSONResponse(w, created, http.StatusCreated)
}
