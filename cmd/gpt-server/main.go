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
		// manager.AddClient(c.ID, enum.GPT_3_5_TURBO, 100)
		c.SetOnMessageReceived(func(c *wsserver.Client, m *wsserver.Message) {
			if m.Action == "ChatAction" {
			}
			gpt, _ := manager.MergeClient(c.ID, enum.GPT_3_5_TURBO, 100)
			if gpt == nil {
				server.Send(c.ID, "You reached max limit")
			}
			res, err := gpt.SendText(m.Message)
			if err != nil {
				log.Println(err)
			}
			log.Println(res)
			server.Send(c.ID, res)
		})
		c.SetOnSettingsReceived(func(c *wsserver.Client, m *wsserver.Message) {
			gpt, _ := manager.MergeClient(c.ID, enum.GPT_3_5_TURBO, 100)
			gpt.SetPrompt(m.Data, nil)
		})
	})
	server.Start("7878")
	select {}
}
