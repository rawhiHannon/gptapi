package chatapp

import (
	"gptapi/internal/idgen"
	"gptapi/internal/openai"
	"gptapi/internal/storage/redis"
	"gptapi/internal/wsserver"
	"gptapi/pkg/api/httpserver"
	"gptapi/pkg/utils"
	"log"
	"time"
)

type ChatApp struct {
	server     *httpserver.HttpServer
	gptManager *openai.GPTManager
	cache      *redis.RedisClient
}

func NewChatApp() *ChatApp {
	h := &ChatApp{}
	h.server = httpserver.NewHttpServer()
	h.init()
	h.initRestAPI()
	return h
}

func (h *ChatApp) init() {
	utils.LoadEnv("")
	h.cache = redis.NewRedisClient("localhost:6379")
	h.gptManager = openai.NewGPTManager(h.cache)
	h.server.SetOnClientRegister(func(c *wsserver.Client) {
		gpt, _ := h.gptManager.GetClient(c.Token)
		gpt.SetRateLimitMsg("Subscripe to talk more")
		c.SetOnMessageReceived(func(c *wsserver.Client, m *wsserver.Message) {
			if m.Action == "ChatAction" {
			}
			gpt, _ := h.gptManager.GetClient(c.Token)
			if gpt == nil {
				return
			}
			res := gpt.SendText(m.Message)
			log.Println(res)
			h.server.Send(c.ID, res)
		})
		c.SetOnSettingsReceived(func(c *wsserver.Client, m *wsserver.Message) {
			gpt, _ := h.gptManager.GetClient(c.Token)
			gpt.SetPrompt(m.Data, nil)
		})
	})
}

func (h *ChatApp) initRestAPI() {
	h.server.RegisterAction("GET", "/generate", h.GenerateToken)
}

func (h *ChatApp) StartAPI(port string) {
	h.server.Start(port)
}

func (h *ChatApp) GenerateToken(params map[string]string, queryString map[string][]string, bodyJson map[string]interface{}) (string, error) {
	token := h.gptManager.GenerateToken("", idgen.NextId(), 10, 4, time.Minute)
	return token, nil
}
