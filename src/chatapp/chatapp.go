package chatapp

import (
	"fmt"
	"gptapi/internal/openai"
	"gptapi/internal/storage/redis"
	"gptapi/internal/uniqid"
	"gptapi/internal/wsserver"
	"gptapi/pkg/api/httpserver"
	"gptapi/pkg/utils"
	"os"
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
	redisHost := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	h.cache = redis.NewRedisClient(fmt.Sprintf(`%s:%s`, redisHost, port))
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
			h.server.Send(c.ID, res)
		})
		c.SetOnSettingsReceived(func(c *wsserver.Client, m *wsserver.Message) {
			gpt, _ := h.gptManager.GetClient(c.Token)
			gpt.SetPrompt(m.Data, nil)
		})
	})
}

func (h *ChatApp) initRestAPI() {
	h.server.RegisterAction("GET", "/generate", h.generateToken)
}

func (h *ChatApp) StartAPI(port string) {
	h.server.Start(port)
}

func (h *ChatApp) generateToken(params map[string]string, queryString map[string][]string, bodyJson map[string]interface{}) (string, error) {
	token := h.gptManager.GenerateToken("", uniqid.NextId(), 10, 4, time.Minute)
	return token, nil
}
