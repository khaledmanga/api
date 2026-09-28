package api

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"api/api/handler"
	"api/api/middleware"
	"api/internal/cache"
	"api/internal/config"
	"api/internal/database"
	"api/internal/repository"
	"api/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.NewDatabase(database.NewMySQLStrategy(&cfg.Database)).Connect()
	if err != nil {
		return fmt.Errorf("connect ent database: %w", err)
	}
	defer db.Close()

	sqlDB, err := database.NewSQLDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("connect sql database: %w", err)
	}
	defer sqlDB.Close()

	cacheClient, err := cache.New(&cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect cache: %w", err)
	}
	defer cacheClient.Close()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			message := "Internal Server Error"
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
				message = fiberErr.Message
			} else {
				log.Printf("request failed: %v", err)
			}
			return c.Status(status).JSON(fiber.Map{
				"error": message,
			})
		},
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete, fiber.MethodOptions},
		AllowHeaders:     []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	userRepository := repository.NewUserRepository(db)
	postRepository := repository.NewPostRepository(sqlDB)
	likeRepository := repository.NewLikeRepository(sqlDB)
	commentRepository := repository.NewCommentRepository(sqlDB)

	authService := service.NewAuthService(userRepository, &cfg.Password)
	postService := service.NewPostService(postRepository, likeRepository)
	commentService := service.NewCommentService(commentRepository, postRepository)
	likeService := service.NewLikeService(likeRepository, postRepository)

	authHandler := handler.NewAuthHandler(authService, cacheClient, cfg.SessionCookieSameSite)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	likeHandler := handler.NewLikeHandler(likeService)

	RegisterRoutes(
		app,
		authHandler,
		postHandler,
		commentHandler,
		likeHandler,
		cacheClient,
		userRepository,
	)

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	if err := app.Listen(addr); err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return nil
}

func RegisterRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	likeHandler *handler.LikeHandler,
	cacheClient *cache.Client,
	userRepository repository.UserRepository,
) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)

	authentication := middleware.Authentication(cacheClient, userRepository)
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
