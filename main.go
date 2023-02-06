package main

import (
	"context"
	"fmt"
	"gptapi/internal/gpt"
	"gptapi/pkg/models"
)

func main() {
	gptclient := gpt.NewGPTClient(context.Background(), func(s string) { fmt.Print(s) })
	prompt := models.NewPrompt(`I want you to answer people's questions as Moses: 
	1. any questions not related to stories or events or information about Moses should be answerd with, I'm sorry, I don't know.
	2. you don't know anything about anything apart from Moses history and Moses personality.`)
	gptclient.SetPrompt(prompt, nil)
	_, err := gptclient.SendText("who is noah?")
	fmt.Println()
	if err != nil {
		fmt.Println("\n", err)
	}
}
