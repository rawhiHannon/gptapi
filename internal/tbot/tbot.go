package tbot

import (
	"fmt"
	"gptapi/internal/openai"
	"gptapi/internal/safe"
	"gptapi/pkg/models"
	"gptapi/pkg/utils"
	"log"
	"os"
	"time"

	"github.com/yanzay/tbot"
)

const (
	DEFAULT_RATE   = time.Minute
	DEFAULT_LIMIT  = 4
	DEFAULT_WINDOW = 10
	DEFAULT_MSG    = "Please Subscribe"
)

type TelegramBot struct {
	gptManager   *openai.GPTManager
	bot          *tbot.Server
	cache        models.CacheManager
	chatMap      safe.SafeMap
	prompt       string
	rateLimitMsg string
	rate         time.Duration
	limit        int
	window       int
}

func NewTelegramBot(cache models.CacheManager) *TelegramBot {
	bot := &TelegramBot{
		cache:        cache,
		rateLimitMsg: DEFAULT_MSG,
		window:       DEFAULT_WINDOW,
		limit:        DEFAULT_LIMIT,
		rate:         DEFAULT_RATE,
	}
	bot.init()
	return bot
}

func (t *TelegramBot) init() {
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
	t.gptManager = openai.NewGPTManager(t.cache)
	t.chatMap = safe.NewSafeMap()
	bot.HandleFunc("{question}", t.questionHandler)
	bot.ListenAndServe()
}

func (t *TelegramBot) getChatKey(chatId int64) string {
	id := fmt.Sprintf(`%d`, chatId)
	token, _ := t.chatMap.Merge(id, func(s string) interface{} {
		return t.gptManager.GenerateToken("bot", uint64(chatId), t.window, t.limit, t.rate)
	})
	return token.(string)
}

func (t *TelegramBot) getChat(chatId int64) openai.IGPTClient {
	token := t.getChatKey(chatId)
	client, exists := t.gptManager.GetClient(token)
	if !exists && client != nil {
		client.SetPrompt(t.prompt, nil)
		client.SetRateLimitMsg(t.rateLimitMsg)
	}
	return client
}

func (t *TelegramBot) questionHandler(m *tbot.Message) {
	question := m.Vars["question"]
	log.Println(question, m.ChatID)
	answer := t.getChat(m.ChatID).SendText(question)
	if answer != "" {
		m.Reply(answer)
	}
}

func (t *TelegramBot) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *TelegramBot) SetRateLimitMsg(msg string) {
	t.rateLimitMsg = msg
}

func (t *TelegramBot) Start() {
	select {}
}
