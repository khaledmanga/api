package services

import (
	"context"

	"api/src/domain"
	"api/src/params/response"
	"api/src/repository"
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

func (s *CommentService) ListComments(ctx context.Context, postID, userID int) ([]response.CommentResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, err
	}
	comments, err := s.commentRepository.ListComments(ctx, postID)
	if err != nil {
		return nil, err
	}
	result := make([]response.CommentResponse, 0, len(comments))
	for i := range comments {
		result = append(result, mapCommentResponse(&comments[i]))
	}
	return result, nil
}

func (s *CommentService) CreateComment(ctx context.Context, postID, authorID int, parentID *int, content string) (*response.CommentResponse, error) {
	comment, err := s.commentRepository.CreateComment(ctx, postID, authorID, parentID, content)
	if err != nil {
		return nil, err
	}
	result := mapCommentResponse(comment)
	return &result, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID, authorID int, content string) error {
	return s.commentRepository.UpdateComment(ctx, commentID, authorID, content)
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID, authorID int) error {
	return s.commentRepository.DeleteComment(ctx, commentID, authorID)
}

func mapCommentResponse(comment *domain.Comment) response.CommentResponse {
	result := response.CommentResponse{
		ID:         comment.ID,
		PostID:     comment.PostID,
		AuthorID:   comment.AuthorID,
		AuthorName: comment.AuthorName,
		ParentID:   comment.ParentID,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
		Replies:    make([]response.CommentResponse, 0, len(comment.Replies)),
	}
	for i := range comment.Replies {
		result.Replies = append(result.Replies, mapCommentResponse(&comment.Replies[i]))
	}
	return result
}
