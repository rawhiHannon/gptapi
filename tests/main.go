package main

import (
	"fmt"
	"gptapi/internal/openai/dalle"
	"os"
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
	// client := dalle.NewClient(apiKey)
	// data, err := client.Generate("a horse cartoon play with football", nil, nil, nil, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(data[0].URL)

	image, _ := os.Open("./files/images/cat.jpg")
	mask, _ := os.Open("./files/images/cat_mask.jpg")
	d := dalle.NewDallE(apiKey, "rawhi", 10*time.Second)
	// list, err := d.GenPhoto("a horse cartoon play with football", 1, "512x512")
	size := dalle.Medium
	n := 1
	responseFormat := "b64_json"
	user := "rawhi"
	list, err := d.Edit("replace the cat with dog", image, mask, &size, &n, &user, &responseFormat)
	fmt.Println(err, list)
}
