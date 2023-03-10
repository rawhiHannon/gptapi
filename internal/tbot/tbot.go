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

type ITelegramBot interface {
	SetPrompt(string)
	SetRateLimitMsg(string)
	Start()
}

type TBot struct {
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
	ready        chan struct{}
}

func NewTBot(botKey string, cache models.CacheManager) ITelegramBot {
	bot := &TBot{
		botKey:       botKey,
		cache:        cache,
		rateLimitMsg: DEFAULT_MSG,
		window:       DEFAULT_WINDOW,
		limit:        DEFAULT_LIMIT,
		rate:         DEFAULT_RATE,
		ready:        make(chan struct{}),
	}
	bot.init(botKey)
	return bot
}

func (t *TBot) init(botKey string) {
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

func (t *TBot) getChatKey(chatId int64) string {
	id := fmt.Sprintf(`%d`, chatId)
	token, _ := t.chatMap.Merge(id, func(s string) interface{} {
		return t.gptManager.GenerateToken("bot", uint64(chatId), t.window, t.limit, t.rate)
	})
	return token.(string)
}

func (t *TBot) getChat(chatId int64) models.IGPTClient {
	token := t.getChatKey(chatId)
	client, exists := t.gptManager.GetClient(token)
	if !exists && client != nil {
		client.SetPrompt(t.prompt, nil)
		client.SetRateLimitMsg(t.rateLimitMsg)
	}
	return client
}

func (t *TBot) questionHandler(m *tbot.Message) {
	question := m.Vars["question"]
	log.Println(question, m.ChatID, m.From)
	answers := t.getChat(m.ChatID).SendText(question)
	for _, answer := range answers {
		if answer.Data != "" {
			if answer.AnswerType == enum.IMAGE_ANSWER {
				m.ReplyPhoto(answer.Data)
			} else {
				m.Reply("Click to Open [URL](http://example.com)")
			}
		}
	}
}

func (t *TBot) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *TBot) SetRateLimitMsg(msg string) {
	t.rateLimitMsg = msg
}

func (t *TBot) Start() {
	t.bot.ListenAndServe()
}
