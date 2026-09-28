package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"api/internal/model"
)

type CommentRepository interface {
	CreateComment(context.Context, int, int, *int, string) (*model.Comment, error)
	ListComments(context.Context, int) ([]model.Comment, error)
	UpdateComment(context.Context, int, int, string) error
	DeleteComment(context.Context, int, int) error
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) CreateComment(ctx context.Context, postID, authorID int, parentID *int, content string) (*model.Comment, error) {
	if parentID != nil {
		var parentPostID int
		if err := r.db.QueryRowContext(ctx, `SELECT post_id FROM comments WHERE id = ?`, *parentID).Scan(&parentPostID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("check parent comment: %w", err)
		}
		if parentPostID != postID {
			return nil, ErrNotFound
		}
	}
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO comments (post_id, author_id, parent_id, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))`,
		postID, authorID, parentID, content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read last insert id: %w", err)
	}
	return r.getComment(ctx, int(id))
}

func (r *commentRepository) ListComments(ctx context.Context, postID int) ([]model.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT comments.id, comments.post_id, comments.author_id, comments.parent_id,
		        comments.content, comments.created_at, comments.updated_at,
		        COALESCE(users.username, '')
		 FROM comments
		 LEFT JOIN users ON users.id = comments.author_id
		 WHERE comments.post_id = ?
		 ORDER BY comments.created_at ASC, comments.id ASC`,
		postID,
	)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	byId := make(map[int]*model.Comment)
	topLevel := make([]model.Comment, 0)
	var all []*model.Comment

	for rows.Next() {
		var comment model.Comment
		var savedParentID sql.NullInt64
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&savedParentID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.AuthorName,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		if savedParentID.Valid {
			parent := int(savedParentID.Int64)
			comment.ParentID = &parent
		}
		comment.Replies = make([]model.Comment, 0)
		copied := comment
		all = append(all, &copied)
		byId[comment.ID] = &copied
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	for _, comment := range all {
		if comment.ParentID == nil {
			topLevel = append(topLevel, *comment)
			continue
		}
		parent, ok := byId[*comment.ParentID]
		if ok {
			parent.Replies = append(parent.Replies, *comment)
		} else {
			topLevel = append(topLevel, *comment)
		}
	}
	return topLevel, nil
}

func (r *commentRepository) getComment(ctx context.Context, id int) (*model.Comment, error) {
	var comment model.Comment
	var savedParentID sql.NullInt64
	if err := r.db.QueryRowContext(ctx,
		`SELECT comments.id, comments.post_id, comments.author_id, comments.parent_id,
		        comments.content, comments.created_at, comments.updated_at,
		        COALESCE(users.username, '')
		 FROM comments
		 LEFT JOIN users ON users.id = comments.author_id
		 WHERE comments.id = ?`,
		id,
	).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&savedParentID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.AuthorName,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan comment by id: %w", err)
	}
	if savedParentID.Valid {
		parent := int(savedParentID.Int64)
		comment.ParentID = &parent
	}
	comment.Replies = make([]model.Comment, 0)
	return &comment, nil
}

func (r *commentRepository) UpdateComment(ctx context.Context, id, authorID int, content string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE comments SET content = ?,
		 updated_at = GREATEST(CURRENT_TIMESTAMP(3), updated_at + INTERVAL 1000 MICROSECOND)
		 WHERE id = ? AND author_id = ?`,
		content, id, authorID,
	)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	if err := requireAffectedRows(result); err != nil {
		return fmt.Errorf("update comment affected rows: %w", err)
	}
	return nil
}

func (r *commentRepository) DeleteComment(ctx context.Context, id, authorID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete comment transaction: %w", err)
	}
	defer tx.Rollback()

	var postID int
	var parentID sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		`SELECT post_id, parent_id FROM comments WHERE id = ? AND author_id = ? FOR UPDATE`,
		id, authorID,
	).Scan(&postID, &parentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("query comment for delete: %w", err)
	}
	var newParentID any
	if parentID.Valid {
		newParentID = int(parentID.Int64)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE comments SET parent_id = ? WHERE post_id = ? AND parent_id = ?`,
		newParentID, postID, id,
	); err != nil {
		return fmt.Errorf("reparent child comments: %w", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE id = ? AND author_id = ?`, id, authorID)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if err := requireAffectedRows(result); err != nil {
		return fmt.Errorf("delete comment affected rows: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete comment: %w", err)
	}
	return nil
}
