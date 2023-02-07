package main

import (
	"gptapi/internal/openai/gpt"
	"gptapi/internal/wsserver"
	"gptapi/pkg/api/httpserver"
	"gptapi/pkg/models"
	"log"
)

const propmt1 = `I want you to answer people's questions as Moses:
1. any questions not related to stories or events or information about Moses should be answerd with, I'm sorry, I don't know.
2. you don't know anything about anything apart from Moses history and Moses personality.`

const propmt2 = `you are a math teacher, rules:
1.any question not related to math you answer with, I know math only :).
2.ant text wich is not question or an order related to previous questions should be answered with, I only can help with math problems :).`

func main() {
	manager := gpt.NewGPTManager()
	server := httpserver.NewHttpServer()
	server.SetOnClientRegister(func(c *wsserver.Client) {
		manager.AddClient(c.ID.String(), nil)
		c.SetOnMessageReceived(func(c *wsserver.Client, m *wsserver.Message) {
			if m.Action == "ChatAction" {

			}
			gpt := manager.GetClient(c.ID.String())
			res, err := gpt.SendText(m.Message)
			if err != nil {
				log.Println(err)
			}
			log.Println(res)
			server.Send(c.ID.String(), res)
		})
		c.SetOnSettingsReceived(func(c *wsserver.Client, m *wsserver.Message) {
			gpt := manager.GetClient(c.ID.String())
			gpt.SetPrompt(models.NewPrompt(m.Data), nil)
		})
	})

	server.Start("7878")
	select {}
}
