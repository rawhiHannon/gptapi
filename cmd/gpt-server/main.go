package main

import (
	"gptapi/internal/openai"
	"gptapi/internal/wsserver"
	"gptapi/pkg/api/httpserver"
	"gptapi/pkg/enum"
	"gptapi/pkg/utils"
	"log"
)

func main() {
	utils.LoadEnv("")
	manager := openai.NewGPTManager()
	server := httpserver.NewHttpServer()
	server.SetOnClientRegister(func(c *wsserver.Client) {
		manager.AddClient(c.ID.String(), enum.GPT_3_5_TURBO, nil)
		c.SetOnMessageReceived(func(c *wsserver.Client, m *wsserver.Message) {
			if m.Action == "ChatAction" {
			}
			gpt := manager.GetClient(c.ID.String())
			if gpt == nil {
				server.Send(c.ID.String(), "You reached max limit")
			}
			res, err := gpt.SendText(m.Message)
			if err != nil {
				log.Println(err)
			}
			log.Println(res)
			server.Send(c.ID.String(), res)
		})
		c.SetOnSettingsReceived(func(c *wsserver.Client, m *wsserver.Message) {
			gpt := manager.GetClient(c.ID.String())
			gpt.SetPrompt(m.Data, nil)
		})
	})
	server.Start("7878")
	select {}
}
