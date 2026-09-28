package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"api/internal/model"
)

type LikeRepository interface {
	LikePost(context.Context, int, int) error
	UnlikePost(context.Context, int, int) error
	GetSummaries(context.Context, []int, int) (map[int]model.LikeSummary, error)
}

type likeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) LikePost(ctx context.Context, postID, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO post_likes (post_id, user_id, created_at)
		 VALUES (?, ?, CURRENT_TIMESTAMP(3))
		 ON DUPLICATE KEY UPDATE post_id = post_likes.post_id`,
		postID, userID,
	)
	if err != nil {
		return fmt.Errorf("like post: %w", err)
	}
	return nil
}

func (r *likeRepository) UnlikePost(ctx context.Context, postID, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`,
		postID, userID,
	)
	if err != nil {
		return fmt.Errorf("unlike post: %w", err)
	}
	return nil
}

func (r *likeRepository) GetSummaries(ctx context.Context, postIDs []int, userID int) (map[int]model.LikeSummary, error) {
	summaries := make(map[int]model.LikeSummary, len(postIDs))
	if len(postIDs) == 0 {
		return summaries, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(postIDs)), ",")
	args := make([]any, 0, len(postIDs)+1)
	args = append(args, userID)
	for _, id := range postIDs {
		args = append(args, id)
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT post_id, COUNT(*), MAX(user_id = ?)
		 FROM post_likes WHERE post_id IN (`+placeholders+`)
		 GROUP BY post_id`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("query like summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var postID int
		var summary model.LikeSummary
		if err := rows.Scan(&postID, &summary.Count, &summary.Liked); err != nil {
			return nil, fmt.Errorf("scan like summary: %w", err)
		}
		summaries[postID] = summary
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate like summaries: %w", err)
	}
	return summaries, nil
}
