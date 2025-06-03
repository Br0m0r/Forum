package notifications

import (
	"database/sql"
	"fmt"
	"forum/db"
	"forum/utils"
	"log"
)

// Create inserts a new notification for the given user.
// commentID may be nil for like/dislike notifications.
func Create(userID, initiatorID, postID int64, commentID *int64, kind string) error {
	log.Printf("notifications.Create: Creating notification with userID=%d, initiatorID=%d, postID=%d, kind=%s",
		userID, initiatorID, postID, kind)

	commentIDStr := "nil"
	if commentID != nil {
		commentIDStr = fmt.Sprintf("%d", *commentID)
	}
	log.Printf("notifications.Create: commentID=%s", commentIDStr)

	_, err := db.Database.Exec(
		`INSERT INTO notifications
         (user_id, initiator_id, kind, post_id, comment_id)
         VALUES (?, ?, ?, ?, ?)`,
		userID, initiatorID, kind, postID, commentID,
	)

	if err != nil {
		log.Printf("notifications.Create: Error creating notification: %v", err)
		return err
	}

	log.Printf("notifications.Create: Successfully created notification for userID=%d", userID)
	return nil
}

// UnreadCount returns the number of unread notifications for a user.
func UnreadCount(userID int64) (int, error) {
	log.Printf("notifications.UnreadCount: Counting unread notifications for userID=%d", userID)

	var count int
	err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`,
		userID,
	).Scan(&count)

	if err != nil {
		log.Printf("notifications.UnreadCount: Error counting unread notifications for userID=%d: %v", userID, err)
		return 0, err
	}

	log.Printf("notifications.UnreadCount: Found %d unread notifications for userID=%d", count, userID)
	return count, nil
}

// List retrieves all notifications for a user, newest first.
func List(userID int64) ([]utils.Notification, error) {
	log.Printf("notifications.List: Retrieving notifications for userID=%d", userID)

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
		log.Printf("notifications.List: Error querying notifications for userID=%d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("notifications.List: Successfully executed query for userID=%d", userID)

	var notifs []utils.Notification
	notifCount := 0

	for rows.Next() {
		var n utils.Notification
		var initiatorUsername string
		var commentID sql.NullInt64

		// Allocate the pointer before scanning into it:
		n.Initiator = &utils.User{}

		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.InitiatorID,
			&n.Kind,
			&n.PostID,
			&commentID,
			&n.IsRead,
			&n.CreatedAt,
			&n.Initiator.ID, // safe now that Initiator != nil
			&initiatorUsername,
		); err != nil {
			log.Printf("notifications.List: Error scanning notification row: %v", err)
			return nil, err
		}

		if commentID.Valid {
			cid := commentID.Int64
			n.CommentID = &cid
			log.Printf("notifications.List: Notification #%d has commentID=%d", n.ID, cid)
		} else {
			log.Printf("notifications.List: Notification #%d has no commentID", n.ID)
		}

		n.Initiator.Username = initiatorUsername
		notifs = append(notifs, n)
		notifCount++

		log.Printf("notifications.List: Processed notification #%d of kind=%s from initiator=%s (ID=%d)",
			n.ID, n.Kind, initiatorUsername, n.InitiatorID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("notifications.List: Error iterating through notifications: %v", err)
		return nil, err
	}

	log.Printf("notifications.List: Successfully retrieved %d notifications for userID=%d", notifCount, userID)
	return notifs, nil
}

// MarkAllRead marks all notifications for a user as read.
func MarkAllRead(userID int64) error {
	log.Printf("notifications.MarkAllRead: Marking all notifications as read for userID=%d", userID)

	result, err := db.Database.Exec(
		`UPDATE notifications SET is_read = 1 WHERE user_id = ?`,
		userID,
	)

	if err != nil {
		log.Printf("notifications.MarkAllRead: Error marking notifications as read for userID=%d: %v", userID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("notifications.MarkAllRead: Error getting rows affected: %v", err)
	} else {
		log.Printf("notifications.MarkAllRead: Marked %d notifications as read for userID=%d",
			rowsAffected, userID)
	}

	return nil
}
