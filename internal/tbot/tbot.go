package tbot

import (
	"gptapi/internal/openai"
	"gptapi/pkg/enum"
	"gptapi/pkg/utils"
	"log"
	"os"

	"github.com/yanzay/tbot"
)

type TelegramBot struct {
	gptManager    *openai.GPTManager
	bot           *tbot.Server
	prompt        string
	maxReachedMsg string
}

func NewTelegramBot(prompt string, maxReachedMsg string) *TelegramBot {
	bot := &TelegramBot{}
	bot.init(prompt, maxReachedMsg)
	return bot
}

func (t *TelegramBot) init(prompt string, maxReachedMsg string) {
	utils.LoadEnv("")
	parmaName := "TELEGRAM_TEST_TOKEN"
	if os.Getenv("ENVIROMENT") == "production" {
		parmaName = "TELEGRAM_TOKEN"
	}
	bot, err := tbot.NewServer(os.Getenv(parmaName))
	if err != nil {
		log.Fatal(err)
	}
	t.bot = bot
	t.gptManager = openai.NewGPTManager()
	t.prompt = prompt
	t.maxReachedMsg = maxReachedMsg
	bot.HandleFunc("{question}", t.questionHandler)
	bot.ListenAndServe()
}

func (t *TelegramBot) getChat(chatId int64) openai.IGPTClient {
	client, exists := t.gptManager.MergeClient(uint64(chatId), enum.GPT_3_5_TURBO, 15)
	if !exists && client != nil {
		client.SetPrompt(t.prompt, nil)
		client.SetMaxReachedMsg(t.maxReachedMsg)
	}
	return client
}

func (t *TelegramBot) questionHandler(m *tbot.Message) {
	question := m.Vars["question"]
	log.Println(question, m.ChatID)
	answer, _ := t.getChat(m.ChatID).SendText(question)
	m.Reply(answer)
}

func (t *TelegramBot) Start() {
	select {}
}
