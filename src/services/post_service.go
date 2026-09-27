package services

import (
	"context"

	"api/src/domain"
	"api/src/params/response"
	"api/src/repository"
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

func (s *PostService) CreatePost(ctx context.Context, authorID int, title, content string) (*response.PostResponse, error) {
	post, err := s.postRepository.CreatePost(ctx, authorID, title, content)
	if err != nil {
		return nil, err
	}
	if err := s.addLikeSummary(ctx, post, authorID); err != nil {
		return nil, err
	}
	result := mapPostResponse(post)
	return &result, nil
}

func (s *PostService) ListPosts(ctx context.Context, userID, page, limit int) ([]response.PostResponse, error) {
	posts, err := s.postRepository.ListPosts(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	return s.mapPostResponses(ctx, posts, userID)
}

func (s *PostService) ListPostsByAuthor(ctx context.Context, authorID, page, limit int) ([]response.PostResponse, error) {
	posts, err := s.postRepository.ListPostsByAuthor(ctx, authorID, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	return s.mapPostResponses(ctx, posts, authorID)
}

func (s *PostService) GetProfileSummary(ctx context.Context, authorID int) (response.ProfileSummaryResponse, error) {
	postCount, likeCount, commentCount, err := s.postRepository.GetAuthorSummary(ctx, authorID)
	if err != nil {
		return response.ProfileSummaryResponse{}, err
	}
	return response.ProfileSummaryResponse{
		PostCount:    postCount,
		LikeCount:    likeCount,
		CommentCount: commentCount,
	}, nil
}

func (s *PostService) mapPostResponses(ctx context.Context, posts []domain.Post, userID int) ([]response.PostResponse, error) {
	postIDs := make([]int, len(posts))
	for i := range posts {
		postIDs[i] = posts[i].ID
	}
	summaries, err := s.likeRepository.GetSummaries(ctx, postIDs, userID)
	if err != nil {
		return nil, err
	}
	result := make([]response.PostResponse, 0, len(posts))
	for i := range posts {
		posts[i].LikeCount = summaries[posts[i].ID].Count
		posts[i].Liked = summaries[posts[i].ID].Liked
		result = append(result, mapPostResponse(&posts[i]))
	}
	return result, nil
}

func (s *PostService) GetPost(ctx context.Context, postID, userID int) (*response.PostResponse, error) {
	post, err := s.postRepository.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if err := s.addLikeSummary(ctx, post, userID); err != nil {
		return nil, err
	}
	result := mapPostResponse(post)
	return &result, nil
}

func (s *PostService) UpdatePost(ctx context.Context, postID, authorID int, title, content string) (*response.PostResponse, error) {
	if err := s.postRepository.UpdatePost(ctx, postID, authorID, title, content); err != nil {
		return nil, err
	}
	return s.GetPost(ctx, postID, authorID)
}

func (s *PostService) DeletePost(ctx context.Context, postID, authorID int) error {
	return s.postRepository.DeletePost(ctx, postID, authorID)
}

func (s *PostService) addLikeSummary(ctx context.Context, post *domain.Post, userID int) error {
	summaries, err := s.likeRepository.GetSummaries(ctx, []int{post.ID}, userID)
	if err != nil {
		return err
	}
	summary := summaries[post.ID]
	post.LikeCount = summary.Count
	post.Liked = summary.Liked
	return nil
}

func mapPostResponse(post *domain.Post) response.PostResponse {
	return response.PostResponse{
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
