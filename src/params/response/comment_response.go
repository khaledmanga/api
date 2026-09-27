package response

import "time"

type CommentResponse struct {
	ID         int               `json:"id"`
	PostID     int               `json:"post_id"`
	AuthorID   int               `json:"author_id"`
	AuthorName string           `json:"author_name"`
	ParentID   *int              `json:"parent_id"`
	Content    string            `json:"content"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Replies    []CommentResponse `json:"replies"`
}
