package repository

import (
	"context"
	"database/sql"
	"strings"

	"api/src/domain"
)

type LikeRepository interface {
	LikePost(context.Context, int, int) error
	UnlikePost(context.Context, int, int) error
	GetSummaries(context.Context, []int, int) (map[int]domain.LikeSummary, error)
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
	return err
}

func (r *likeRepository) UnlikePost(ctx context.Context, postID, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`,
		postID, userID,
	)
	return err
}

func (r *likeRepository) GetSummaries(ctx context.Context, postIDs []int, userID int) (map[int]domain.LikeSummary, error) {
	summaries := make(map[int]domain.LikeSummary, len(postIDs))
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
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID int
		var summary domain.LikeSummary
		if err := rows.Scan(&postID, &summary.Count, &summary.Liked); err != nil {
			return nil, err
		}
		summaries[postID] = summary
	}
	return summaries, rows.Err()
}
