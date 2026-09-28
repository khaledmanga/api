package handler

import (
	"strings"
	"unicode/utf8"

	"api/internal/model"
	"api/internal/service"

	"github.com/gofiber/fiber/v3"
)

const maxCommentContentLength = 5000

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) ListComments(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	comments, err := h.commentService.ListComments(c.Context(), postID, userID)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(comments)
}

func (h *CommentHandler) CreateComment(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	var params model.CreateCommentRequest
	if err := c.Bind().Body(&params); err != nil ||
		strings.TrimSpace(params.Content) == "" ||
		utf8.RuneCountInString(params.Content) > maxCommentContentLength ||
		params.ParentID != nil && *params.ParentID < 1 {
		return fiber.ErrBadRequest
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	comment, err := h.commentService.CreateComment(c.Context(), postID, userID, params.ParentID, strings.TrimSpace(params.Content))
	if err != nil {
		return resourceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(comment)
}

func (h *CommentHandler) UpdateComment(c fiber.Ctx) error {
	commentID, err := positiveID(c.Params("commentID"))
	if err != nil {
		return err
	}
	var params model.UpdateCommentRequest
	if err := c.Bind().Body(&params); err != nil ||
		strings.TrimSpace(params.Content) == "" ||
		utf8.RuneCountInString(params.Content) > maxCommentContentLength {
		return fiber.ErrBadRequest
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if err := h.commentService.UpdateComment(c.Context(), commentID, userID, strings.TrimSpace(params.Content)); err != nil {
		return resourceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CommentHandler) DeleteComment(c fiber.Ctx) error {
	commentID, err := positiveID(c.Params("commentID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if err := h.commentService.DeleteComment(c.Context(), commentID, userID); err != nil {
		return resourceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
