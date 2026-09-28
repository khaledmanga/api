package main

import (
	"log"

	"api/api"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return api.Run()
}


