package chatapp

import (
	"gptapi/pkg/api/httpserver"
	"log"
)

type ChatApp struct {
	httServer *httpserver.HttpServer
}

func NewChatApp() *ChatApp {
	c := &ChatApp{}
	c.httServer = httpserver.NewHttpServer()
	return c
}

func (h *ChatApp) AddNode(params map[string]string, queryString map[string][]string, bodyJson map[string]interface{}) (string, error) {
	log.Println("REST WORKS POST")
	return "", nil
}
