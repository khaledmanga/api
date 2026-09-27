package repository

import (
	"context"
	"database/sql"
	"errors"

	"api/src/domain"
)

type CommentRepository interface {
	ListComments(context.Context, int) ([]domain.Comment, error)
	CreateComment(context.Context, int, int, *int, string) (*domain.Comment, error)
	UpdateComment(context.Context, int, int, string) error
	DeleteComment(context.Context, int, int) error
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) ListComments(ctx context.Context, postID int) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT comments.id, comments.post_id, comments.author_id, comments.parent_id,
		        comments.content, comments.created_at, comments.updated_at,
		        COALESCE(users.username, '')
		 FROM comments
		 LEFT JOIN users ON users.id = comments.author_id
		 WHERE comments.post_id = ?
		 ORDER BY comments.created_at, comments.id`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make(map[int]*domain.Comment)
	children := make(map[int][]int)
	roots := make([]int, 0)
	orderedIDs := make([]int, 0)
	for rows.Next() {
		var comment domain.Comment
		var parentID sql.NullInt64
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&parentID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.AuthorName,
		); err != nil {
			return nil, err
		}
		if parentID.Valid {
			id := int(parentID.Int64)
			comment.ParentID = &id
		}
		nodes[comment.ID] = &comment
		orderedIDs = append(orderedIDs, comment.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range orderedIDs {
		comment := nodes[id]
		if comment.ParentID == nil {
			roots = append(roots, id)
		} else if _, exists := nodes[*comment.ParentID]; exists {
			children[*comment.ParentID] = append(children[*comment.ParentID], id)
		} else {
			roots = append(roots, id)
		}
	}

	var build func(int) domain.Comment
	build = func(id int) domain.Comment {
		comment := *nodes[id]
		comment.Replies = make([]domain.Comment, 0, len(children[id]))
		for _, childID := range children[id] {
			comment.Replies = append(comment.Replies, build(childID))
		}
		return comment
	}
	comments := make([]domain.Comment, 0, len(roots))
	for _, id := range roots {
		comments = append(comments, build(id))
	}
	return comments, nil
}

func (r *commentRepository) CreateComment(ctx context.Context, postID, authorID int, parentID *int, content string) (*domain.Comment, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM posts WHERE id = ?`, postID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if parentID != nil {
		if err := r.db.QueryRowContext(ctx,
			`SELECT id FROM comments WHERE id = ? AND post_id = ?`,
			*parentID, postID,
		).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
	}
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO comments (post_id, author_id, parent_id, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))`,
		postID, authorID, parentID, content,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	var comment domain.Comment
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
		return nil, err
	}
	if savedParentID.Valid {
		parent := int(savedParentID.Int64)
		comment.ParentID = &parent
	}
	comment.Replies = make([]domain.Comment, 0)
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
		return err
	}
	return requireAffectedRows(result)
}

func (r *commentRepository) DeleteComment(ctx context.Context, id, authorID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
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
		return err
	}
	var newParentID any
	if parentID.Valid {
		newParentID = int(parentID.Int64)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE comments SET parent_id = ? WHERE post_id = ? AND parent_id = ?`,
		newParentID, postID, id,
	); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE id = ? AND author_id = ?`, id, authorID)
	if err != nil {
		return err
	}
	if err := requireAffectedRows(result); err != nil {
		return err
	}
	return tx.Commit()
}
