package service

import (
	"context"
	"fmt"

	"api/internal/model"
	"api/internal/repository"
)

type CommentService struct {
	commentRepository repository.CommentRepository
	postRepository    repository.PostRepository
}

func NewCommentService(commentRepository repository.CommentRepository, postRepository repository.PostRepository) *CommentService {
	return &CommentService{
		commentRepository: commentRepository,
		postRepository:    postRepository,
	}
}

func (s *CommentService) ListComments(ctx context.Context, postID, userID int) ([]model.CommentResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, fmt.Errorf("verify post exists for comments: %w", err)
	}
	comments, err := s.commentRepository.ListComments(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	result := make([]model.CommentResponse, 0, len(comments))
	for i := range comments {
		result = append(result, mapCommentResponse(&comments[i]))
	}
	return result, nil
}

func (s *CommentService) CreateComment(ctx context.Context, postID, authorID int, parentID *int, content string) (*model.CommentResponse, error) {
	comment, err := s.commentRepository.CreateComment(ctx, postID, authorID, parentID, content)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	result := mapCommentResponse(comment)
	return &result, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID, authorID int, content string) error {
	if err := s.commentRepository.UpdateComment(ctx, commentID, authorID, content); err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID, authorID int) error {
	if err := s.commentRepository.DeleteComment(ctx, commentID, authorID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

func mapCommentResponse(comment *model.Comment) model.CommentResponse {
	result := model.CommentResponse{
		ID:         comment.ID,
		PostID:     comment.PostID,
		AuthorID:   comment.AuthorID,
		AuthorName: comment.AuthorName,
		ParentID:   comment.ParentID,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
		Replies:    make([]model.CommentResponse, 0, len(comment.Replies)),
	}
	for i := range comment.Replies {
		result.Replies = append(result.Replies, mapCommentResponse(&comment.Replies[i]))
	}
	return result
}
