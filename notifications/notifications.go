package notifications

import (
	"database/sql"
	"forum/db"
	"forum/utils"
	"log"
)

// Create inserts a new notification for the given user.
func Create(userID, initiatorID, postID int64, commentID *int64, kind string) error {
	_, err := db.Database.Exec(
		`INSERT INTO notifications
         (user_id, initiator_id, kind, post_id, comment_id)
         VALUES (?, ?, ?, ?, ?)`,
		userID, initiatorID, kind, postID, commentID,
	)
	if err != nil {
		log.Printf("notifications.Create: %v", err)
	}
	return err
}

// UnreadCount returns the number of unread notifications for a user.
func UnreadCount(userID int64) (int, error) {
	var count int
	err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`,
		userID,
	).Scan(&count)
	if err != nil {
		log.Printf("notifications.UnreadCount: %v", err)
	}
	return count, err
}

// List retrieves all notifications for a user, newest first.
func List(userID int64) ([]utils.Notification, error) {
	rows, err := db.Database.Query(
		`SELECT n.id, n.user_id, n.initiator_id, n.kind, n.post_id, n.comment_id,
                n.is_read, n.created_at,
                u.id, u.username
           FROM notifications n
           JOIN users u ON u.id = n.initiator_id
          WHERE n.user_id = ?
          ORDER BY n.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []utils.Notification
	for rows.Next() {
		var n utils.Notification
		var initiatorUsername string
		var commentID sql.NullInt64

		n.Initiator = &utils.User{}

		if err := rows.Scan(
			&n.ID, &n.UserID, &n.InitiatorID, &n.Kind, &n.PostID, &commentID,
			&n.IsRead, &n.CreatedAt, &n.Initiator.ID, &initiatorUsername,
		); err != nil {
			log.Printf("notifications.List: error scanning row: %v", err)
			return nil, err
		}

		if commentID.Valid {
			cid := commentID.Int64
			n.CommentID = &cid
		}

		n.Initiator.Username = initiatorUsername
		notifs = append(notifs, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notifs, nil
}

// MarkAllRead marks all notifications for a user as read.
func MarkAllRead(userID int64) error {
	_, err := db.Database.Exec(
		`UPDATE notifications SET is_read = 1 WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		log.Printf("notifications.MarkAllRead: %v", err)
	}
	return err
}
