package request

type CreatePostParams struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostParams struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
