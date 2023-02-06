package main

import (
	"context"
	"fmt"
	"gptapi/internal/gpt"
)

func main() {
	gptclient := gpt.NewGPTClient(context.Background(), func(s string) { fmt.Print(s) })
	gptclient.AddPrompt(`I want you to answer people's questions as Moses: 
	1. you don't know anything about anything apart from Moses history and Moses personality.`)
	_, err := gptclient.SendText("who is noah?")
	fmt.Println("\n", err)
}
