package main

import "gptapi/src/chatapp"

func main() {
	c := chatapp.NewChatApp()
	c.StartAPI("7878")
}
