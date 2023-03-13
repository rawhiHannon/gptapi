package tbot

import (
	"fmt"
	"gptapi/internal/openai"
	"gptapi/internal/safe"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

type TeleBot struct {
	gptManager   *openai.GPTManager
	botKey       string
	bot          *telebot.Bot
	cache        models.CacheManager
	chatMap      safe.SafeMap
	prompt       string
	rateLimitMsg string
	rate         time.Duration
	limit        int
	window       int
	ready        chan struct{}
}

func NewTeleBot(botKey string, cache models.CacheManager) ITelegramBot {
	bot := &TeleBot{
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

func (t *TeleBot) init(botKey string) {
	t.botKey = botKey
	pref := telebot.Settings{
		Token:  t.botKey,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}
	t.bot = b
	t.gptManager = openai.NewGPTManager(t.cache)
	t.chatMap = safe.NewSafeMap()
	t.bot.Handle(telebot.OnText, t.questionHandler)
}

func (t *TeleBot) getChatKey(chatId int64) string {
	id := fmt.Sprintf(`%d`, chatId)
	token, _ := t.chatMap.Merge(id, func(s string) interface{} {
		return t.gptManager.GenerateToken("bot", uint64(chatId), t.window, t.limit, t.rate)
	})
	return token.(string)
}

func (t *TeleBot) getChat(chatId int64) models.IGPTClient {
	token := t.getChatKey(chatId)
	client, exists := t.gptManager.GetClient(token)
	if !exists && client != nil {
		client.SetPrompt(t.prompt, nil)
		client.SetRateLimitMsg(t.rateLimitMsg)
	}
	return client
}

func (t *TeleBot) questionHandler(c telebot.Context) error {
	var (
		user     = c.Sender()
		question = c.Text()
	)
	log.Println(question, user.ID)
	t.bot.Notify(user, telebot.Typing)
	answers := t.getChat(user.ID).SendText(question)
	var err error
	for _, answer := range answers {
		if answer.Data != "" {
			if answer.AnswerType == enum.IMAGE_ANSWER {
				photo := &telebot.Photo{File: telebot.FromURL(answer.Data)}
				err = c.Send(photo)
			} else {
				err = c.Send(answer.Data)
			}
		}
	}
	return err
}

func (t *TeleBot) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *TeleBot) SetRateLimitMsg(msg string) {
	t.rateLimitMsg = msg
}

func (t *TeleBot) Start() {
	t.bot.Start()
}
