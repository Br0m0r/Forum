package likes

import (
	"database/sql"
	"forum/db"
)

// ToggleLike creates or updates a like/dislike for this user+post.
func ToggleLike(userID, postID int64, isLike bool) error {
	// Check if the user already voted on this post
	var exists bool
	var existingIsLike bool

	err := db.Database.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = ? AND post_id = ?), is_like FROM post_likes WHERE user_id = ? AND post_id = ?",
		userID, postID, userID, postID,
	).Scan(&exists, &existingIsLike)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if exists {
		if existingIsLike == isLike {
			// Same vote again → remove (toggle off)
			_, err = db.Database.Exec(
				"DELETE FROM post_likes WHERE user_id = ? AND post_id = ?",
				userID, postID,
			)
			return err
		} else {
			// Changed vote → update
			_, err = db.Database.Exec(
				"UPDATE post_likes SET is_like = ? WHERE user_id = ? AND post_id = ?",
				isLike, userID, postID,
			)
			return err
		}
	} else {
		// New vote → insert
		_, err = db.Database.Exec(
			"INSERT INTO post_likes (user_id, post_id, is_like) VALUES (?, ?, ?)",
			userID, postID, isLike,
		)
		return err
	}
}

// CountLikes returns (likesCount, dislikesCount, error).
func CountLikes(postID int64) (int, int, error) {
	var likes, dislikes int
	if err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND is_like = 1`,
		postID,
	).Scan(&likes); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND is_like = 0`,
		postID,
	).Scan(&dislikes); err != nil && err != sql.ErrNoRows {
		return likes, 0, err
	}
	return likes, dislikes, nil
}

func CountCommentLikes(commentID int64) (int, int, error) {
	var likes, dislikes int
	if err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = 1`,
		commentID,
	).Scan(&likes); err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err := db.Database.QueryRow(
		`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = 0`,
		commentID,
	).Scan(&dislikes); err != nil && err != sql.ErrNoRows {
		return likes, 0, err
	}
	return likes, dislikes, nil
}

func ToggleCommentLike(userID, commentID int64, isLike bool) error {
	var existingIsLike bool
	err := db.Database.QueryRow(
		"SELECT is_like FROM comment_likes WHERE user_id = ? AND comment_id = ?",
		userID, commentID,
	).Scan(&existingIsLike)

	if err == sql.ErrNoRows {
		// no prior vote → insert new
		_, err = db.Database.Exec(
			"INSERT INTO comment_likes (user_id, comment_id, is_like) VALUES (?, ?, ?)",
			userID, commentID, isLike,
		)
		return err
	} else if err != nil {
		return err
	}

	if existingIsLike == isLike {
		// same vote again → remove (toggle off)
		_, err = db.Database.Exec(
			"DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?",
			userID, commentID,
		)
		return err
	}

	// different vote → update to new value
	_, err = db.Database.Exec(
		"UPDATE comment_likes SET is_like = ? WHERE user_id = ? AND comment_id = ?",
		isLike, userID, commentID,
	)
	return err
}

// GetUserPostVote returns 1 if liked, -1 if disliked, 0 if no vote.
func GetUserPostVote(userID, postID int64) (int, error) {
	var isLike sql.NullBool
	err := db.Database.QueryRow(
		"SELECT is_like FROM post_likes WHERE user_id = ? AND post_id = ?",
		userID, postID,
	).Scan(&isLike)
	if err == sql.ErrNoRows {
		return 0, nil // no vote
	}
	if err != nil {
		return 0, err
	}
	if isLike.Valid {
		if isLike.Bool {
			return 1, nil
		}
		return -1, nil
	}
	return 0, nil
}

// GetUserCommentVote returns 1 if liked, -1 if disliked, 0 if no vote.
func GetUserCommentVote(userID, commentID int64) (int, error) {
	var isLike sql.NullBool
	err := db.Database.QueryRow(
		"SELECT is_like FROM comment_likes WHERE user_id = ? AND comment_id = ?",
		userID, commentID,
	).Scan(&isLike)
	if err == sql.ErrNoRows {
		return 0, nil // no vote
	}
	if err != nil {
		return 0, err
	}
	if isLike.Valid {
		if isLike.Bool {
			return 1, nil
		}
		return -1, nil
	}
	return 0, nil
}
