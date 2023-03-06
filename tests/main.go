package main

import (
	"fmt"
	"gptapi/internal/openai"
	"time"
)

func main() {
	// gptclient := gpt.NewGPTClient(context.Background(), func(s string) { fmt.Print(s) })
	// prompt := models.NewPrompt(`I want you to answer people's questions as Moses:
	// 1. any questions not related to stories or events or information about Moses should be answerd with, I'm sorry, I don't know.
	// 2. you don't know anything about anything apart from Moses history and Moses personality.`)
	// gptclient.SetPrompt(prompt, nil)
	// _, err := gptclient.SendText("who is noah?")
	// fmt.Println()
	// if err != nil {
	// 	fmt.Println("\n", err)
	// }

	apiKey := "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"
	d := openai.NewDallE(apiKey, "rawhi", 10*time.Second)

	list, err := d.GenPhoto(`
	Film still, extreme wide shot of an
	elephant alone on the savannah,
	extreme long shot
	`, 1, "1024x1024")

	// image, _ := os.Open("./files/images/cat.jpg")
	// mask, _ := os.Open("./files/images/cat_mask.jpg")
	// size := openai.Medium
	// n := 1
	// responseFormat := "b64_json"
	// user := "rawhi"
	// list, err := d.Edit("replace the cat with dog", image, mask, &size, &n, &user, &responseFormat)
	fmt.Println(err, list)
}
