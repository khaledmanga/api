package request

type CreateCommentParams struct {
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id"`
}

type UpdateCommentParams struct {
	Content string `json:"content"`
}
