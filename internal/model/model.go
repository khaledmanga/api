package model

import "time"

type UserRole int

const (
	ADMIN UserRole = iota
	USER
)

func (r UserRole) String() string {
	return []string{"admin", "user"}[r]
}

type UserState int

const (
	PENDING UserState = iota
	ACTIVE
	INACTIVE
	BANNED
)

func (s UserState) String() string {
	return []string{"pending", "active", "inactive", "banned"}[s]
}

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

type Post struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	AuthorID     int       `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	Liked        bool      `json:"liked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Like struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type LikeSummary struct {
	Count int  `json:"count"`
	Liked bool `json:"liked"`
}

type Comment struct {
	ID         int       `json:"id"`
	PostID     int       `json:"post_id"`
	AuthorID   int       `json:"author_id"`
	AuthorName string    `json:"author_name"`
	ParentID   *int      `json:"parent_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Replies    []Comment `json:"replies"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type AuthResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type PostResponse struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	AuthorID     int       `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	Liked        bool      `json:"liked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProfileSummaryResponse struct {
	PostCount    int `json:"post_count"`
	LikeCount    int `json:"like_count"`
	CommentCount int `json:"comment_count"`
}

type LikeResponse struct {
	Liked bool `json:"liked"`
}

type CommentResponse struct {
	ID         int               `json:"id"`
	PostID     int               `json:"post_id"`
	AuthorID   int               `json:"author_id"`
	AuthorName string            `json:"author_name"`
	ParentID   *int              `json:"parent_id"`
	Content    string            `json:"content"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Replies    []CommentResponse `json:"replies"`
}
