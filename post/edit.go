package post

import (
	"fmt"
	"forum/db"
)

// EditPostByID updates the content of a post in the posts table.
func EditPostByID(postID int64, content string) error {
	// ← corrected table name here
	stmt, err := db.Database.Prepare("UPDATE posts SET content = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare error: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(content, postID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no post found with id %d", postID)
	}

	return nil
}

// EditCommentByID updates the content of a comment in the comments table.
func EditCommentByID(commentID int64, content string) error {
	stmt, err := db.Database.Prepare("UPDATE comments SET content = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare error: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(content, commentID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no comment found with id %d", commentID)
	}

	return nil
}
