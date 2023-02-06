package main

import (
	"gptapi/pkg/api/httpserver"
)

func main() {
	// manager := gpt.NewGPTManager()
	server := httpserver.NewHttpServer()
	server.Start("7878")

	select {}
}
