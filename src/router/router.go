package router

import (
	"api/src/config"
	"api/src/handler"
	"api/src/middleware"
	"api/src/repository"

	"github.com/gofiber/fiber/v3"
)

func NewRouter(
	app *fiber.App,
	authHandler handler.AuthHandler,
	postHandler handler.PostHandler,
	commentHandler handler.CommentHandler,
	likeHandler handler.LikeHandler,
	redisClient *config.RedisClient,
	userRepository repository.UserRepository,
) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)

	authentication := middleware.Authentication(redisClient, userRepository)
	auth.Get("/me", authentication, authHandler.Me)
	auth.Post("/logout", authentication, authHandler.Logout)
	posts := api.Group("/posts", authentication)
	posts.Get("/mine/summary", postHandler.GetMyProfileSummary)
	posts.Get("/mine", postHandler.ListMyPosts)
	posts.Get("/", postHandler.ListPosts)
	posts.Post("/", postHandler.CreatePost)
	posts.Get("/:postID", postHandler.GetPost)
	posts.Put("/:postID", postHandler.UpdatePost)
	posts.Delete("/:postID", postHandler.DeletePost)
	posts.Get("/:postID/comments", commentHandler.ListComments)
	posts.Post("/:postID/comments", commentHandler.CreateComment)
	posts.Post("/:postID/likes", likeHandler.LikePost)
	posts.Delete("/:postID/likes", likeHandler.UnlikePost)

	comments := api.Group("/comments", authentication)
	comments.Put("/:commentID", commentHandler.UpdateComment)
	comments.Delete("/:commentID", commentHandler.DeleteComment)
}
