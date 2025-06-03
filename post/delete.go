package post

import (
	"fmt"
	"forum/db"
	"log"
)

// / DeletePostByID deletes a post by its ID, returning an error if nothing was removed.
func DeletePostByID(postID int) error {
	log.Printf("DeletePostByID: Starting deletion of post_id=%d", postID)

	// Begin a transaction to ensure all operations succeed or fail together
	tx, err := db.Database.Begin()
	if err != nil {
		log.Printf("DeletePostByID: Transaction begin error for post_id=%d: %v", postID, err)
		return fmt.Errorf("transaction begin error: %w", err)
	}
	log.Printf("DeletePostByID: Transaction started for post_id=%d", postID)

	// First delete notifications related to comments on this post
	log.Printf("DeletePostByID: Deleting notifications for comments on post_id=%d", postID)
	_, err = tx.Exec("DELETE FROM notifications WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", postID)
	if err != nil {
		log.Printf("DeletePostByID: Failed to delete comment notifications for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("delete comment notifications error: %w", err)
	}
	log.Printf("DeletePostByID: Successfully deleted comment notifications for post_id=%d", postID)

	// Delete notifications related to the post itself
	log.Printf("DeletePostByID: Deleting notifications for post_id=%d", postID)
	_, err = tx.Exec("DELETE FROM notifications WHERE post_id = ?", postID)
	if err != nil {
		log.Printf("DeletePostByID: Failed to delete post notifications for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("delete post notifications error: %w", err)
	}
	log.Printf("DeletePostByID: Successfully deleted post notifications for post_id=%d", postID)

	// Delete all comment likes/votes for comments on this post
	log.Printf("DeletePostByID: Trying to delete comment_likes for post_id=%d", postID)
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", postID)
	if err != nil {
		log.Printf("DeletePostByID: Failed with comment_likes, trying comment_votes for post_id=%d: %v", postID, err)
		// Try comment_votes if comment_likes fails
		_, err = tx.Exec("DELETE FROM comment_votes WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", postID)
		if err != nil {
			log.Printf("DeletePostByID: Failed to delete comment votes for post_id=%d: %v", postID, err)
			tx.Rollback()
			return fmt.Errorf("delete comment votes error: %w", err)
		}
		log.Printf("DeletePostByID: Successfully deleted comment_votes for post_id=%d", postID)
	} else {
		log.Printf("DeletePostByID: Successfully deleted comment_likes for post_id=%d", postID)
	}

	// Delete all comments associated with the post
	log.Printf("DeletePostByID: Deleting comments for post_id=%d", postID)
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", postID)
	if err != nil {
		log.Printf("DeletePostByID: Failed to delete comments for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("delete comments error: %w", err)
	}
	log.Printf("DeletePostByID: Successfully deleted comments for post_id=%d", postID)

	// Delete all post likes/votes
	log.Printf("DeletePostByID: Trying to delete post_likes for post_id=%d", postID)
	_, err = tx.Exec("DELETE FROM post_likes WHERE post_id = ?", postID)
	if err != nil {
		log.Printf("DeletePostByID: Failed with post_likes, trying post_votes for post_id=%d: %v", postID, err)
		// Try post_votes if post_likes fails
		_, err = tx.Exec("DELETE FROM post_votes WHERE post_id = ?", postID)
		if err != nil {
			log.Printf("DeletePostByID: Failed to delete post votes for post_id=%d: %v", postID, err)
			tx.Rollback()
			return fmt.Errorf("delete post votes error: %w", err)
		}
		log.Printf("DeletePostByID: Successfully deleted post_votes for post_id=%d", postID)
	} else {
		log.Printf("DeletePostByID: Successfully deleted post_likes for post_id=%d", postID)
	}

	// Now delete the post itself
	log.Printf("DeletePostByID: Preparing statement to delete post_id=%d", postID)
	stmt, err := tx.Prepare("DELETE FROM posts WHERE id = ?")
	if err != nil {
		log.Printf("DeletePostByID: Failed to prepare statement for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("prepare error: %w", err)
	}
	defer stmt.Close()
	log.Printf("DeletePostByID: Statement prepared for post_id=%d", postID)

	log.Printf("DeletePostByID: Executing delete for post_id=%d", postID)
	res, err := stmt.Exec(postID)
	if err != nil {
		log.Printf("DeletePostByID: Execution failed for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("exec error: %w", err)
	}
	log.Printf("DeletePostByID: Delete executed for post_id=%d", postID)

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("DeletePostByID: Failed to get rows affected for post_id=%d: %v", postID, err)
		tx.Rollback()
		return fmt.Errorf("rows affected error: %w", err)
	}
	log.Printf("DeletePostByID: Delete affected %d rows for post_id=%d", rowsAffected, postID)

	if rowsAffected == 0 {
		log.Printf("DeletePostByID: No post found with id=%d", postID)
		tx.Rollback()
		return fmt.Errorf("no post found with id %d", postID)
	}

	// Commit the transaction
	log.Printf("DeletePostByID: Committing transaction for post_id=%d", postID)
	if err = tx.Commit(); err != nil {
		log.Printf("DeletePostByID: Failed to commit transaction for post_id=%d: %v", postID, err)
		return fmt.Errorf("transaction commit error: %w", err)
	}
	log.Printf("DeletePostByID: Successfully committed transaction for post_id=%d", postID)

	log.Printf("DeletePostByID: Post with id=%d successfully deleted", postID)
	return nil
}

// DeleteCommentByID deletes a comment by its ID, returning an error if nothing was removed.
func DeleteCommentByID(commentID int) error {
	log.Printf("DeleteCommentByID: Starting deletion of comment_id=%d", commentID)

	// Begin a transaction
	tx, err := db.Database.Begin()
	if err != nil {
		log.Printf("DeleteCommentByID: Transaction begin error for comment_id=%d: %v", commentID, err)
		return fmt.Errorf("transaction begin error: %w", err)
	}
	log.Printf("DeleteCommentByID: Transaction started for comment_id=%d", commentID)

	// First delete notifications associated with this comment
	log.Printf("DeleteCommentByID: Deleting notifications for comment_id=%d", commentID)
	_, err = tx.Exec("DELETE FROM notifications WHERE comment_id = ?", commentID)
	if err != nil {
		log.Printf("DeleteCommentByID: Failed to delete notifications for comment_id=%d: %v", commentID, err)
		tx.Rollback()
		return fmt.Errorf("delete notifications error: %w", err)
	}
	log.Printf("DeleteCommentByID: Successfully deleted notifications for comment_id=%d", commentID)

	// Delete all likes/votes associated with this comment
	log.Printf("DeleteCommentByID: Trying to delete comment_likes for comment_id=%d", commentID)
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id = ?", commentID)
	if err != nil {
		log.Printf("DeleteCommentByID: Failed with comment_likes, trying comment_votes for comment_id=%d: %v", commentID, err)
		// Try comment_votes if comment_likes fails
		_, err = tx.Exec("DELETE FROM comment_votes WHERE comment_id = ?", commentID)
		if err != nil {
			log.Printf("DeleteCommentByID: Failed to delete comment votes for comment_id=%d: %v", commentID, err)
			tx.Rollback()
			return fmt.Errorf("delete comment votes error: %w", err)
		}
		log.Printf("DeleteCommentByID: Successfully deleted comment_votes for comment_id=%d", commentID)
	} else {
		log.Printf("DeleteCommentByID: Successfully deleted comment_likes for comment_id=%d", commentID)
	}

	// Now delete the comment itself
	log.Printf("DeleteCommentByID: Preparing statement to delete comment_id=%d", commentID)
	stmt, err := tx.Prepare("DELETE FROM comments WHERE id = ?")
	if err != nil {
		log.Printf("DeleteCommentByID: Failed to prepare statement for comment_id=%d: %v", commentID, err)
		tx.Rollback()
		return fmt.Errorf("prepare error: %w", err)
	}
	defer stmt.Close()
	log.Printf("DeleteCommentByID: Statement prepared for comment_id=%d", commentID)

	log.Printf("DeleteCommentByID: Executing delete for comment_id=%d", commentID)
	res, err := stmt.Exec(commentID)
	if err != nil {
		log.Printf("DeleteCommentByID: Execution failed for comment_id=%d: %v", commentID, err)
		tx.Rollback()
		return fmt.Errorf("exec error: %w", err)
	}
	log.Printf("DeleteCommentByID: Delete executed for comment_id=%d", commentID)

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("DeleteCommentByID: Failed to get rows affected for comment_id=%d: %v", commentID, err)
		tx.Rollback()
		return fmt.Errorf("rows affected error: %w", err)
	}
	log.Printf("DeleteCommentByID: Delete affected %d rows for comment_id=%d", rowsAffected, commentID)

	if rowsAffected == 0 {
		log.Printf("DeleteCommentByID: No comment found with id=%d", commentID)
		tx.Rollback()
		return fmt.Errorf("no comment found with id %d", commentID)
	}

	// Commit the transaction
	log.Printf("DeleteCommentByID: Committing transaction for comment_id=%d", commentID)
	if err = tx.Commit(); err != nil {
		log.Printf("DeleteCommentByID: Failed to commit transaction for comment_id=%d: %v", commentID, err)
		return fmt.Errorf("transaction commit error: %w", err)
	}
	log.Printf("DeleteCommentByID: Successfully committed transaction for comment_id=%d", commentID)

	log.Printf("DeleteCommentByID: Comment with id=%d successfully deleted", commentID)
	return nil
}
