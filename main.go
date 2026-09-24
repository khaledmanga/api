package main

import (
	"api/src/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(
				fiber.Map{
					"error": err.Error()
				}
			)
		}
	})
}
