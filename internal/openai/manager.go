package openai

import (
	"fmt"
	"gptapi/internal/jwt"
	"gptapi/internal/limiter"
	"gptapi/internal/safe"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type IGPTClient interface {
	SetPrompt(string, []string)
	SetRateLimitMsg(string)
	SendText(string) string
}

type GPTManager struct {
	clients      map[uint64]IGPTClient
	cache        models.CacheManager
	tokenManager *jwt.JWT
	clientsMap   safe.SafeMap
	apiKeys      []string
	limiter      *limiter.RedisRateLimiter
	jhash        *JumpHash
	mu           *sync.RWMutex
}

func NewGPTManager(cache models.CacheManager) *GPTManager {
	m := &GPTManager{
		clients: make(map[uint64]IGPTClient),
		mu:      new(sync.RWMutex),
	}
	m.cache = cache
	m.clientsMap = safe.NewSafeMap()
	m.tokenManager = jwt.New(cache, os.Getenv("GPT_JWT_SECRET"))
	m.loadApiKeys()
	m.jhash = newJumpHash(len(m.apiKeys), 1)
	return m
}

func (m *GPTManager) loadApiKeys() {
	keys, ok := os.LookupEnv("GPT_API_KEY")
	if !ok || len(keys) == 0 {
		panic("OpenAI: api keys missing")
	}
	m.apiKeys = strings.Split(keys, ",")
}

func (m *GPTManager) getApiKey(key uint64) string {
	pos := m.jhash.get(key)
	return m.apiKeys[pos]
}

func (m *GPTManager) decodeToken(token string) *jwt.TokenPayload {
	payload, err := m.tokenManager.ValidateToken(token)
	if err != nil {
		return nil
	}
	return payload
}

func (m *GPTManager) GenerateToken(identifier string, accessId uint64, window, limit int, rate time.Duration) string {
	payload := map[string]interface{}{
		"limit":  limit,
		"rate":   rate,
		"window": window,
	}
	token, err := m.tokenManager.CreateToken(identifier, accessId, payload)
	if err != nil {
		return ""
	}
	return token.Token
}

func (m *GPTManager) GetClient(token string) (IGPTClient, bool) {
	payload := m.decodeToken(token)
	if payload == nil {
		log.Println(token)
		return nil, false
	}
	data := payload.Data
	limit := int(data["limit"].(float64))
	window := int(data["window"].(float64))
	rate := time.Duration(int64(data["rate"].(float64)))
	c, exists := m.clientsMap.Merge(fmt.Sprintf(`%d`, payload.AccessId), func(s string) interface{} {
		c := CreateNewGPTClient(payload.AccessId, m.getApiKey(payload.AccessId), enum.GPT_3_5_TURBO, window, limit, rate)
		return c
	})
	return c.(IGPTClient), exists
}
