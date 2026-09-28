package main

import (
	"log"

	"api/api"
)

func main() {
	if err := api.Run(); err != nil {
		log.Fatal(err)
	}
}

