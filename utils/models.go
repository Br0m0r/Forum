package utils

import (
	"fmt"
	"time"
)

type GroupedPosts struct {
	Authored []Post
	Reacted  []Post
}

type TemplateData struct {
	Username         string
	Posts            []Post
	Authored         []Post
	Reacted          []Post
	Post             Post
	Comments         []Comment
	Categories       []Category
	SelectedCategory int64
	Notifications    []Notification
	NotifCount       int
}

type User struct {
	ID       int64
	Username string
	Email    string
	Password *string
	// the * is useful in situations where the password may be nullable (i.e., it can be nil or not set)
	// strings cant be null, but pointers to a string can
	CreatedAt time.Time
}
type Post struct {
	ID             int64
	Title          string
	Content        string
	UserID         int64
	UserName       string
	CategoryName   string
	Likes_count    int // now dynamically populated
	Dislikes_count int // now dynamically populated
	Comments_count int // new
	CreatedAt      time.Time
	PostType       string
	Image          string
	Comment        []Comment
	CategoryIDs    []int64  // NEW
	CategoryNames  []string // NEW
	UserVote       int
}

type Comment struct {
	ID             int64
	Content        string
	PostID         int64
	UserID         int64
	UserName       string
	Likes_count    int
	Dislikes_count int
	CreatedAt      time.Time
	UserVote       int
}

type Category struct {
	ID   int64
	Name string
}

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

type Notification struct {
	ID          int64
	UserID      int64
	InitiatorID int64
	Kind        string
	PostID      int64
	CommentID   *int64
	IsRead      bool
	CreatedAt   time.Time
	Initiator   *User
}

func (n Notification) DisplayText() string {
	switch n.Kind {
	case "like":
		return fmt.Sprintf("%s liked your post", n.Initiator.Username)
	case "dislike":
		return fmt.Sprintf("%s disliked your post", n.Initiator.Username)
	case "comment":
		return fmt.Sprintf("%s commented on your post", n.Initiator.Username)
	case "comment_like":
		return fmt.Sprintf("%s liked your comment", n.Initiator.Username)
	case "comment_dislike":
		return fmt.Sprintf("%s disliked your comment", n.Initiator.Username)
	default:
		return "You have a new notification"
	}
}
