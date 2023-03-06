package tbot

import (
	"fmt"
	"gptapi/internal/openai"
	"gptapi/pkg/enum"
	"gptapi/pkg/utils"
	"log"
	"os"

	"github.com/yanzay/tbot"
)

type TelegramBot struct {
	gptManager *openai.GPTManager
	bot        *tbot.Server
	prompt     string
}

func NewTelegramBot(prompt string) *TelegramBot {
	bot := &TelegramBot{}
	bot.init(prompt)
	return bot
}

func (t *TelegramBot) init(prompt string) {
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
	bot.HandleFunc("{question}", t.questionHandler)
	bot.ListenAndServe()
}

func (t *TelegramBot) getChat(chatId int64) openai.IGPTClient {
	client := t.gptManager.GetClient(fmt.Sprintf(`%d`, chatId))
	if client == nil {
		client = t.gptManager.AddClient(os.Getenv("GPT_API_KEY"), enum.GPT_3_5_TURBO, 5)
		client.SetPrompt(t.prompt, nil)
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
