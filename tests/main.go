package main

import (
	"fmt"
	"gptapi/internal/storage/redis"
	"gptapi/internal/tbot"
	"gptapi/pkg/utils"
	"os"
)

var empty string

var rule3 string = `
Any request for generating a photo, you should answer with {photo description in english}, without any other text.
If the user ask for modifications, you should answer with {photo updated description in english}, without any other text.
Any question not related to generating photos, you should answer I only can help with making photos.
Here's an example to give you an idea:
User: Can you generate a photo of a cat jumping?
Answer format: {a photo of a cat jumping}
User: I don't want the cat inside the house.
Answer format: {a photo of a cat jumping outside the house}
`

func main() {
	utils.LoadEnv("")
	redisHost := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	botKey := os.Getenv("TELEGRAM_TEST_TOKEN")
	r := redis.NewRedisClient(fmt.Sprintf(`%s:%s`, redisHost, port))
	bot := tbot.NewTelegramBot(botKey, r)
	bot.SetPrompt(rule3)
	bot.Start()

	// de := openai.NewDallE(os.Getenv("GPT_API_KEY"), 1, 10, 0)
	// res, err := de.GenPhoto("A lion dancing dabka may depict a lion standing on its hind legs with its front legs bent, holding hands together while performing the traditional Arab folk dance called 'Dabke' The lion may be wearing traditional clothing that is commonly seen in the Middle East, such as a shemagh or keffiyeh. Its facial expression may show a joyful and celebratory mood while dancing.", 1, "512x512")
	// log.Println(err, res)
}
