package openai

import (
	"fmt"
	"gptapi/internal/jwt"
	"gptapi/internal/limiter"
	"gptapi/internal/safe"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"os"
	"strings"
	"sync"
	"time"
)

type GPTManager struct {
	clients      map[uint64]models.IGPTClient
	cache        models.CacheManager
	gptType      enum.GPTType
	tokenManager *jwt.JWT
	clientsMap   safe.SafeMap
	apiKeys      []string
	limiter      *limiter.RedisRateLimiter
	jhash        *JumpHash
	mu           *sync.RWMutex
}

func NewGPTManager(cache models.CacheManager) *GPTManager {
	m := &GPTManager{
		clients: make(map[uint64]models.IGPTClient),
		gptType: enum.GPT_3_5_TURBO,
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

func (m *GPTManager) SetType(gptType enum.GPTType) {
	m.gptType = gptType
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
		"window": window,
		"limit":  limit,
		"rate":   rate,
	}
	token, err := m.tokenManager.CreateToken(identifier, accessId, payload)
	if err != nil {
		return ""
	}
	return token.Token
}

func (m *GPTManager) GetClient(token string) (models.IGPTClient, bool) {
	payload := m.decodeToken(token)
	if payload == nil {
		return nil, false
	}
	window := int(payload.Data["window"].(float64))
	limit := int(payload.Data["limit"].(float64))
	rate := time.Duration(int64(payload.Data["rate"].(float64)))
	apiKey := m.getApiKey(payload.AccessId)
	id := fmt.Sprintf(`%d`, payload.AccessId)
	c, exists := m.clientsMap.Merge(id, func(s string) interface{} {
		c := CreateNewGPTClient(payload.AccessId, apiKey, m.gptType, window, limit, rate)
		return c
	})
	return c.(models.IGPTClient), exists
}
