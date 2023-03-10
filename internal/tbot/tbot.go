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
	DEFUALT_RATE   = time.Minute
	DEFAULT_LIMIT  = 4
	DEFAULT_WINDOW = 10
)

type TelegramBot struct {
	gptManager    *openai.GPTManager
	bot           *tbot.Server
	cache         models.CacheManager
	chatMap       safe.SafeMap
	prompt        string
	maxReachedMsg string
	rate          time.Duration
	limit         int
	window        int
}

func NewTelegramBot(cache models.CacheManager, prompt string, maxReachedMsg string, window, limit int, rate time.Duration) *TelegramBot {
	bot := &TelegramBot{}
	bot.init(cache, prompt, maxReachedMsg, window, limit, rate)
	return bot
}

func (t *TelegramBot) init(cache models.CacheManager, prompt string, maxReachedMsg string, window, limit int, rate time.Duration) {
	utils.LoadEnv("")
	parmaName := "TELEGRAM_TEST_TOKEN"
	if os.Getenv("ENVIROMENT") == "production" {
		parmaName = "TELEGRAM_TOKEN"
	}
	bot, err := tbot.NewServer(os.Getenv(parmaName))
	if err != nil {
		log.Fatal(err)
	}
	t.cache = cache
	t.bot = bot
	t.window = window
	t.limit = limit
	t.rate = rate
	if t.window == 0 {
		t.window = DEFAULT_WINDOW
	}
	if t.limit == 0 {
		t.limit = DEFAULT_LIMIT
	}
	if t.rate == 0 {
		t.rate = DEFUALT_RATE
	}
	t.gptManager = openai.NewGPTManager(t.cache)
	t.chatMap = safe.NewSafeMap()
	t.prompt = prompt
	t.maxReachedMsg = maxReachedMsg
	bot.HandleFunc("{question}", t.questionHandler)
	bot.ListenAndServe()
}

func (t *TelegramBot) getChatKey(chatId int64) string {
	token, _ := t.chatMap.Merge(fmt.Sprintf(`%d`, chatId), func(s string) interface{} {
		return t.gptManager.GenerateToken("bot", DEFAULT_WINDOW, DEFAULT_LIMIT, DEFUALT_RATE)
	})
	return token.(string)
}

func (t *TelegramBot) getChat(chatId int64) openai.IGPTClient {
	token := t.getChatKey(chatId)
	client, exists := t.gptManager.GetClient(token)
	if !exists && client != nil {
		client.SetPrompt(t.prompt, nil)
		client.SetRateLimitMsg(t.maxReachedMsg)
	}
	return client
}

func (t *TelegramBot) questionHandler(m *tbot.Message) {
	question := m.Vars["question"]
	log.Println(question, m.ChatID)
	answer, _ := t.getChat(m.ChatID).SendText(question)
	if answer != "" {
		m.Reply(answer)
	}
}

func (t *TelegramBot) Start() {
	select {}
}
