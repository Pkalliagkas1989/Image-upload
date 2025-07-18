package models

import "time"

// Notification represents a user notification
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ActorID   string    `json:"actor_id"`
	PostID    *string   `json:"post_id,omitempty"`
	CommentID *string   `json:"comment_id,omitempty"`
	Action    string    `json:"action"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationWithActor includes the username of the actor
type NotificationWithActor struct {
	Notification
	ActorUsername string `json:"actor_username"`
}
