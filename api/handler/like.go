package handler

import (
	"api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type LikeHandler struct {
	likeService *service.LikeService
}

func NewLikeHandler(likeService *service.LikeService) *LikeHandler {
	return &LikeHandler{likeService: likeService}
}

func (h *LikeHandler) LikePost(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	result, err := h.likeService.LikePost(c.Context(), postID, userID)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(result)
}

func (h *LikeHandler) UnlikePost(c fiber.Ctx) error {
	postID, err := positiveID(c.Params("postID"))
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	result, err := h.likeService.UnlikePost(c.Context(), postID, userID)
	if err != nil {
		return resourceError(err)
	}
	return c.JSON(result)
}
