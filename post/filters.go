package post

import (
	"database/sql"
	"fmt"
	"forum/db"
	"forum/likes"
	"forum/utils"
)

func FilteredPosts(categoryID int64) ([]utils.Post, error) {
	var posts []utils.Post
	var img sql.NullString

	query := `
        SELECT p.id, p.title, p.content, p.user_id, p.user_name, p.created_at,image
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
        WHERE pc.category_id = ?
        ORDER BY p.created_at DESC
    `

	rows, err := db.Database.Query(query, categoryID)
	if err != nil {
		fmt.Println("Error retrieving posts by category:", err)
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post utils.Post

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.UserName,
			&post.CreatedAt,
			&img,
		)
		if err != nil {
			fmt.Println("Error scanning post row:", err)
			continue
		}
		if img.Valid {
			post.Image = img.String

		}

		// Add likes/dislikes
		likesCount, dislikesCount, err := likes.CountLikes(post.ID)
		if err != nil {
			fmt.Printf("Error getting likes for post %d: %v\n", post.ID, err)
		}
		post.Likes_count = likesCount
		post.Dislikes_count = dislikesCount

		// Fetch categories for each post
		categoryQuery := `
            SELECT c.id, c.name
            FROM categories c
            INNER JOIN post_categories pc ON c.id = pc.category_id
            WHERE pc.post_id = ?
        `
		catRows, err := db.Database.Query(categoryQuery, post.ID)
		if err != nil {
			fmt.Printf("Error retrieving categories for post %d: %v\n", post.ID, err)
			continue
		}

		for catRows.Next() {
			var id int64
			var name string
			if err := catRows.Scan(&id, &name); err != nil {
				fmt.Printf("Error scanning category for post %d: %v\n", post.ID, err)
				continue
			}
			post.CategoryIDs = append(post.CategoryIDs, id)
			post.CategoryNames = append(post.CategoryNames, name)
		}
		catRows.Close()

		posts = append(posts, post)
	}

	return posts, nil
}
