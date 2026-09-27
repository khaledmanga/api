package main

import (
	"errors"
	"log"
	"net"
	"strconv"

	"api/src/config"
	"api/src/handler"
	"api/src/mapper"
	"api/src/repository"
	"api/src/router"
	"api/src/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	db, err := config.NewDatabase(config.NewMySQLStrategy(&cfg.Database)).Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	sqlDB, err := config.NewSQLDB(&cfg.Database)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	redisClient, err := config.NewRedisClient(&cfg.Redis)
	if err != nil {
		return err
	}
	defer redisClient.Close()

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
	authService := services.NewAuthService(
		userRepository,
		&cfg.Password,
	)
	authHandler := handler.NewAuthHandler(
		mapper.NewUserMapper(),
		authService,
		redisClient,
		cfg.SessionCookieSameSite,
	)
	postRepository := repository.NewPostRepository(sqlDB)
	likeRepository := repository.NewLikeRepository(sqlDB)
	commentRepository := repository.NewCommentRepository(sqlDB)

	postService := services.NewPostService(postRepository, likeRepository)
	commentService := services.NewCommentService(commentRepository, postRepository)
	likeService := services.NewLikeService(likeRepository, postRepository)

	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	likeHandler := handler.NewLikeHandler(likeService)

	router.NewRouter(
		app,
		*authHandler,
		*postHandler,
		*commentHandler,
		*likeHandler,
		redisClient,
		userRepository,
	)

	return app.Listen(net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port)))
}
