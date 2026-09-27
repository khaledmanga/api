package domain

import "time"

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
