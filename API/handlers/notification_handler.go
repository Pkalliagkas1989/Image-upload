package handlers

import (
	"encoding/json"
	"net/http"

	"forum/middleware"
	"forum/repository"
	"forum/utils"
)

// NotificationHandler handles user notifications
type NotificationHandler struct {
	Repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{Repo: repo}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	notifs, err := h.Repo.GetByUser(user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}
	utils.JSONResponse(w, notifs, http.StatusOK)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
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
		ID string `json:"id"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}
	var err error
	if req.ID == "" {
		err = h.Repo.MarkAllRead(user.ID)
	} else {
		err = h.Repo.MarkRead(user.ID, req.ID)
	}
	if err != nil {
		utils.ErrorResponse(w, "Failed to mark read", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// ID can be provided either as a query param or in the JSON body.
	id := r.URL.Query().Get("id")
	var err error

	if id == "" && r.Body != nil {
		var req struct {
			ID *string `json:"id"`
		}
		// Ignore decode errors as body may be empty
		json.NewDecoder(r.Body).Decode(&req)
		if req.ID != nil {
			id = *req.ID
		}
	}

	// Some clients may send the string "null" when no ID is supplied
	if id == "" || id == "null" {
		err = h.Repo.DeleteAll(user.ID)
	} else {
		err = h.Repo.Delete(user.ID, id)
	}
	if err != nil {
		utils.ErrorResponse(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
