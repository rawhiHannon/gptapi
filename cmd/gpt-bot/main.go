package main

import "gptapi/internal/tbot"

func main() {
	bot := tbot.NewTelegramBot(`rules: 
	. if user ask "do you love me" in any language, you should answer: tell me your name (in the asked language).
	. if the name is malak then answer I love you so much and tell her why as it is your fiancee.
	`)
	bot.Start()
}
