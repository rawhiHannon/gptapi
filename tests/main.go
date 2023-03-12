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
I want you to act as a music app for recommendations.
rules:
. you only answer questions about music videos and links.
. when asked about songs and video clips try get the link from youtube.
`

func main() {
	utils.LoadEnv("")
	redisHost := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	botKey := os.Getenv("TELEGRAM_RAWHI_BOT")
	r := redis.NewRedisClient(fmt.Sprintf(`%s:%s`, redisHost, port))
	bot := tbot.NewTelegramBot(botKey, r)
	bot.SetPrompt(empty)
	bot.Start()

	// de := openai.NewDallE(os.Getenv("GPT_API_KEY"), 1, 10, 0)
	// res, err := de.GenPhoto("حصان بلعب فطبول", 1, "512x512")
	// log.Println(err, res)
}
