package service

import (
	"context"
	"fmt"

	"api/internal/model"
	"api/internal/repository"
)

type LikeService struct {
	likeRepository repository.LikeRepository
	postRepository repository.PostRepository
}

func NewLikeService(likeRepository repository.LikeRepository, postRepository repository.PostRepository) *LikeService {
	return &LikeService{
		likeRepository: likeRepository,
		postRepository: postRepository,
	}
}

func (s *LikeService) LikePost(ctx context.Context, postID, userID int) (*model.LikeResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, fmt.Errorf("verify post exists before like: %w", err)
	}
	if err := s.likeRepository.LikePost(ctx, postID, userID); err != nil {
		return nil, fmt.Errorf("like post in repository: %w", err)
	}
	return &model.LikeResponse{Liked: true}, nil
}

func (s *LikeService) UnlikePost(ctx context.Context, postID, userID int) (*model.LikeResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, fmt.Errorf("verify post exists before unlike: %w", err)
	}
	if err := s.likeRepository.UnlikePost(ctx, postID, userID); err != nil {
		return nil, fmt.Errorf("unlike post in repository: %w", err)
	}
	return &model.LikeResponse{Liked: false}, nil
}
