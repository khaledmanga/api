package services

import (
	"context"

	"api/src/params/response"
	"api/src/repository"
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

func (s *LikeService) LikePost(ctx context.Context, postID, userID int) (*response.LikeResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, err
	}
	if err := s.likeRepository.LikePost(ctx, postID, userID); err != nil {
		return nil, err
	}
	return &response.LikeResponse{Liked: true}, nil
}

func (s *LikeService) UnlikePost(ctx context.Context, postID, userID int) (*response.LikeResponse, error) {
	if _, err := s.postRepository.GetPost(ctx, postID); err != nil {
		return nil, err
	}
	if err := s.likeRepository.UnlikePost(ctx, postID, userID); err != nil {
		return nil, err
	}
	return &response.LikeResponse{Liked: false}, nil
}
