package response

import "time"

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
