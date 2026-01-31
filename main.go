package main

import (
	"log"

	"github.com/X0Ken/openai-gateway/cmd/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
