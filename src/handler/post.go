package handler

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"api/src/params/request"
	"api/src/services"

	"github.com/gofiber/fiber/v3"
)

const (
	maxPostTitleLength   = 200
	maxPostContentLength = 10000
	defaultPostPageSize  = 20
	maxPostPageSize      = 100
)

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) CreatePost(c fiber.Ctx) error {
	var params request.CreatePostParams
	if err := c.Bind().Body(&params); err != nil ||
		strings.TrimSpace(params.Content) == "" ||
		utf8.RuneCountInString(params.Title) > maxPostTitleLength ||
		utf8.RuneCountInString(params.Content) > maxPostContentLength {
		return fiber.ErrBadRequest
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	post, err := h.postService.CreatePost(
		c.Context(),
		userID,
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Content),
	)
	if err != nil {
		return resourceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) ListPosts(c fiber.Ctx) error {
	page, limit, err := postPagination(c)
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	posts, err := h.postService.ListPosts(c.Context(), userID, page, limit)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(posts)
}

func (h *PostHandler) ListMyPosts(c fiber.Ctx) error {
	page, limit, err := postPagination(c)
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	posts, err := h.postService.ListPostsByAuthor(c.Context(), userID, page, limit)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(posts)
}

func (h *PostHandler) GetMyProfileSummary(c fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	summary, err := h.postService.GetProfileSummary(c.Context(), userID)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(summary)
}

func (h *PostHandler) GetPost(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	post, err := h.postService.GetPost(c.Context(), postID, userID)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(post)
}

func (h *PostHandler) UpdatePost(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	var params request.UpdatePostParams
	if err := c.Bind().Body(&params); err != nil ||
		strings.TrimSpace(params.Content) == "" ||
		utf8.RuneCountInString(params.Title) > maxPostTitleLength ||
		utf8.RuneCountInString(params.Content) > maxPostContentLength {
		return fiber.ErrBadRequest
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	post, err := h.postService.UpdatePost(
		c.Context(),
		postID,
		userID,
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Content),
	)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(post)
}

func postPagination(c fiber.Ctx) (int, int, error) {
	page := 1
	limit := defaultPostPageSize
	if raw := c.Query("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fiber.ErrBadRequest
		}
		page = parsed
	}
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxPostPageSize {
			return 0, 0, fiber.ErrBadRequest
		}
		limit = parsed
	}
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/limit {
		return 0, 0, fiber.ErrBadRequest
	}
	return page, limit, nil
}

func (h *PostHandler) DeletePost(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if err := h.postService.DeletePost(c.Context(), postID, userID); err != nil {
		return resourceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
