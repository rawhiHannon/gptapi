package main

import "gptapi/internal/tbot"

func main() {
	bot := tbot.NewTelegramBot("")
	bot.Start()
}
