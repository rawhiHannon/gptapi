package tbot

import (
	"fmt"
	"gptapi/internal/openai"
	"gptapi/internal/safe"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"log"
	"time"

	"github.com/yanzay/tbot"
)

const (
	DEFAULT_RATE   = time.Minute
	DEFAULT_LIMIT  = 40
	DEFAULT_WINDOW = 10
	DEFAULT_MSG    = "ONHOLD"
)

type TelegramBot struct {
	gptManager   *openai.GPTManager
	botKey       string
	bot          *tbot.Server
	cache        models.CacheManager
	chatMap      safe.SafeMap
	prompt       string
	rateLimitMsg string
	rate         time.Duration
	limit        int
	window       int
}

func NewTelegramBot(botKey string, cache models.CacheManager) *TelegramBot {
	bot := &TelegramBot{
		botKey:       botKey,
		cache:        cache,
		rateLimitMsg: DEFAULT_MSG,
		window:       DEFAULT_WINDOW,
		limit:        DEFAULT_LIMIT,
		rate:         DEFAULT_RATE,
	}
	bot.init(botKey)
	return bot
}

func (t *TelegramBot) init(botKey string) {
	t.botKey = botKey
	bot, err := tbot.NewServer(t.botKey)
	if err != nil {
		log.Fatal(err)
	}
	t.bot = bot
	t.gptManager = openai.NewGPTManager(t.cache)
	t.chatMap = safe.NewSafeMap()
	t.bot.HandleFunc("{question}", t.questionHandler)
}

func (t *TelegramBot) getChatKey(chatId int64) string {
	id := fmt.Sprintf(`%d`, chatId)
	token, _ := t.chatMap.Merge(id, func(s string) interface{} {
		return t.gptManager.GenerateToken("bot", uint64(chatId), t.window, t.limit, t.rate)
	})
	return token.(string)
}

func (t *TelegramBot) getChat(chatId int64) models.IGPTClient {
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
	log.Println(question, m.ChatID, m.From)
	answer, anserType := t.getChat(m.ChatID).SendText(question)
	if answer != "" {
		if anserType == enum.IMAGE_ANSWER {
			m.ReplyPhoto(answer)
		} else {
			m.Reply(answer)
		}
	}
}

func (t *TelegramBot) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *TelegramBot) SetRateLimitMsg(msg string) {
	t.rateLimitMsg = msg
}

func (t *TelegramBot) Start() {
	t.bot.ListenAndServe()
}
