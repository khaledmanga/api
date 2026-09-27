package repository

import (
	"context"
	"database/sql"
	"errors"

	"api/src/domain"
)

type PostRepository interface {
	CreatePost(context.Context, int, string, string) (*domain.Post, error)
	ListPosts(context.Context, int, int) ([]domain.Post, error)
	ListPostsByAuthor(context.Context, int, int, int) ([]domain.Post, error)
	GetAuthorSummary(context.Context, int) (int, int, int, error)
	GetPost(context.Context, int) (*domain.Post, error)
	UpdatePost(context.Context, int, int, string, string) error
	DeletePost(context.Context, int, int) error
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func scanPost(row interface{ Scan(...any) error }) (*domain.Post, error) {
	var post domain.Post
	if err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.AuthorName,
		&post.CommentCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

const postSelect = `
	SELECT p.id, p.title, p.content, p.author_id, p.created_at, p.updated_at,
	       COALESCE(authors.username, ''),
	       COALESCE(comment_counts.comment_count, 0)
	FROM posts p
	LEFT JOIN users AS authors ON authors.id = p.author_id
	LEFT JOIN (
		SELECT post_id, COUNT(*) AS comment_count
		FROM comments
		GROUP BY post_id
	) AS comment_counts ON comment_counts.post_id = p.id`

func (r *postRepository) CreatePost(ctx context.Context, authorID int, title, content string) (*domain.Post, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO posts (title, content, author_id, created_at, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))`,
		title, content, authorID,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetPost(ctx, int(id))
}

func (r *postRepository) ListPosts(ctx context.Context, offset, limit int) ([]domain.Post, error) {
	rows, err := r.db.QueryContext(ctx, postSelect+` ORDER BY p.created_at DESC, p.id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, *post)
	}
	return posts, rows.Err()
}

func (r *postRepository) ListPostsByAuthor(ctx context.Context, authorID, offset, limit int) ([]domain.Post, error) {
	rows, err := r.db.QueryContext(ctx,
		postSelect+` WHERE p.author_id = ? ORDER BY p.created_at DESC, p.id DESC LIMIT ? OFFSET ?`,
		authorID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, *post)
	}
	return posts, rows.Err()
}

func (r *postRepository) GetAuthorSummary(ctx context.Context, authorID int) (int, int, int, error) {
	var postCount, likeCount, commentCount int
	err := r.db.QueryRowContext(ctx,
		`SELECT
			(SELECT COUNT(*) FROM posts WHERE author_id = ?),
			(SELECT COUNT(*) FROM post_likes
			 JOIN posts ON posts.id = post_likes.post_id
			 WHERE posts.author_id = ?),
			(SELECT COUNT(*) FROM comments
			 JOIN posts ON posts.id = comments.post_id
			 WHERE posts.author_id = ?)`,
		authorID, authorID, authorID,
	).Scan(&postCount, &likeCount, &commentCount)
	return postCount, likeCount, commentCount, err
}

func (r *postRepository) GetPost(ctx context.Context, id int) (*domain.Post, error) {
	return scanPost(r.db.QueryRowContext(ctx, postSelect+` WHERE p.id = ?`, id))
}

func (r *postRepository) UpdatePost(ctx context.Context, id, authorID int, title, content string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE posts SET title = ?, content = ?,
		 updated_at = GREATEST(CURRENT_TIMESTAMP(3), updated_at + INTERVAL 1000 MICROSECOND)
		 WHERE id = ? AND author_id = ?`,
		title, content, id, authorID,
	)
	if err != nil {
		return err
	}
	return requireAffectedRows(result)
}

func (r *postRepository) DeletePost(ctx context.Context, id, authorID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingID int
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM posts WHERE id = ? AND author_id = ? FOR UPDATE`,
		id, authorID,
	).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE post_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_likes WHERE post_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM posts WHERE id = ? AND author_id = ?`, id, authorID); err != nil {
		return err
	}
	return tx.Commit()
}
