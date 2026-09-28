package service

import (
	"context"
	"fmt"

	"api/internal/model"
	"api/internal/repository"
)

type PostService struct {
	postRepository repository.PostRepository
	likeRepository repository.LikeRepository
}

func NewPostService(postRepository repository.PostRepository, likeRepository repository.LikeRepository) *PostService {
	return &PostService{
		postRepository: postRepository,
		likeRepository: likeRepository,
	}
}

func (s *PostService) CreatePost(ctx context.Context, authorID int, title, content string) (*model.PostResponse, error) {
	post, err := s.postRepository.CreatePost(ctx, authorID, title, content)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	if err := s.addLikeSummary(ctx, post, authorID); err != nil {
		return nil, fmt.Errorf("add like summary: %w", err)
	}
	res := mapPostResponse(post)
	return &res, nil
}

func (s *PostService) ListPosts(ctx context.Context, userID, page, limit int) ([]model.PostResponse, error) {
	posts, err := s.postRepository.ListPosts(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	return s.mapPostResponses(ctx, posts, userID)
}

func (s *PostService) ListPostsByAuthor(ctx context.Context, authorID, page, limit int) ([]model.PostResponse, error) {
	posts, err := s.postRepository.ListPostsByAuthor(ctx, authorID, (page-1)*limit, limit)
	if err != nil {
		return nil, fmt.Errorf("list author posts: %w", err)
	}
	return s.mapPostResponses(ctx, posts, authorID)
}

func (s *PostService) GetProfileSummary(ctx context.Context, authorID int) (model.ProfileSummaryResponse, error) {
	postCount, likeCount, commentCount, err := s.postRepository.GetAuthorSummary(ctx, authorID)
	if err != nil {
		return model.ProfileSummaryResponse{}, fmt.Errorf("get author summary: %w", err)
	}
	return model.ProfileSummaryResponse{
		PostCount:    postCount,
		LikeCount:    likeCount,
		CommentCount: commentCount,
	}, nil
}

func (s *PostService) mapPostResponses(ctx context.Context, posts []model.Post, userID int) ([]model.PostResponse, error) {
	ids := make([]int, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}
	summaries, err := s.likeRepository.GetSummaries(ctx, ids, userID)
	if err != nil {
		return nil, fmt.Errorf("get summaries for posts: %w", err)
	}
	result := make([]model.PostResponse, 0, len(posts))
	for _, p := range posts {
		if summary, ok := summaries[p.ID]; ok {
			p.LikeCount = summary.Count
			p.Liked = summary.Liked
		}
		result = append(result, mapPostResponse(&p))
	}
	return result, nil
}

func (s *PostService) GetPost(ctx context.Context, postID, userID int) (*model.PostResponse, error) {
	post, err := s.postRepository.GetPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	if err := s.addLikeSummary(ctx, post, userID); err != nil {
		return nil, fmt.Errorf("add like summary: %w", err)
	}
	res := mapPostResponse(post)
	return &res, nil
}

func (s *PostService) UpdatePost(ctx context.Context, postID, authorID int, title, content string) (*model.PostResponse, error) {
	if err := s.postRepository.UpdatePost(ctx, postID, authorID, title, content); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	return s.GetPost(ctx, postID, authorID)
}

func (s *PostService) DeletePost(ctx context.Context, postID, authorID int) error {
	if err := s.postRepository.DeletePost(ctx, postID, authorID); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func (s *PostService) addLikeSummary(ctx context.Context, post *model.Post, userID int) error {
	summaries, err := s.likeRepository.GetSummaries(ctx, []int{post.ID}, userID)
	if err != nil {
		return fmt.Errorf("get like summary: %w", err)
	}
	if summary, ok := summaries[post.ID]; ok {
		post.LikeCount = summary.Count
		post.Liked = summary.Liked
	}
	return nil
}

func mapPostResponse(post *model.Post) model.PostResponse {
	return model.PostResponse{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		AuthorID:     post.AuthorID,
		AuthorName:   post.AuthorName,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		Liked:        post.Liked,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}
