package repository

import (
	"database/sql"
	"time"

	"forum/models"
	"forum/utils"
)

// NotificationRepository handles DB operations for notifications
type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n models.Notification) error {
	n.ID = utils.GenerateUUID()
	n.CreatedAt = time.Now()
	isRead := 0
	_, err := r.db.Exec(`INSERT INTO notifications (notification_id, user_id, actor_id, post_id, comment_id, action, is_read, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.UserID, n.ActorID, n.PostID, n.CommentID, n.Action, isRead, n.CreatedAt)
	return err
}

func (r *NotificationRepository) GetByUser(userID string) ([]models.NotificationWithActor, error) {
	rows, err := r.db.Query(`SELECT n.notification_id, n.user_id, n.actor_id, u.username, n.post_id, n.comment_id, n.action, n.is_read, n.created_at FROM notifications n JOIN user u ON n.actor_id = u.user_id WHERE n.user_id = ? ORDER BY n.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.NotificationWithActor
	for rows.Next() {
		var n models.NotificationWithActor
		var postID, commentID sql.NullString
		err := rows.Scan(&n.ID, &n.UserID, &n.ActorID, &n.ActorUsername, &postID, &commentID, &n.Action, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		if postID.Valid {
			n.PostID = &postID.String
		}
		if commentID.Valid {
			n.CommentID = &commentID.String
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *NotificationRepository) MarkAllRead(userID string) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ?`, userID)
	return err
}

func (r *NotificationRepository) MarkRead(userID, id string) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = 1 WHERE notification_id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *NotificationRepository) Delete(userID, id string) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE notification_id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *NotificationRepository) DeleteAll(userID string) error {
	_, err := r.db.Exec(`DELETE FROM notifications WHERE user_id = ?`, userID)
	return err
}
