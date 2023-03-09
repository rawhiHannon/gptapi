package openai

import (
	"fmt"
	"gptapi/internal/limiter"
	"gptapi/pkg/enum"
	"os"
	"strings"
	"sync"
	"time"
)

type IGPTClient interface {
	SetPrompt(string, []string)
	SetMaxReachedMsg(string)
	SendText(string) (string, error)
}

type GPTManager struct {
	clients    map[uint64]IGPTClient
	clientsMap SafeMap
	apiKeys    []string
	limiter    *limiter.RedisRateLimiter
	jhash      *JumpHash
	mu         *sync.RWMutex
}

func NewGPTManager() *GPTManager {
	m := &GPTManager{
		clients: make(map[uint64]IGPTClient),
		mu:      new(sync.RWMutex),
	}
	m.clientsMap = NewSafeMap()
	m.limiter = limiter.NewRedisRateLimiter("GPT:LIMITER:", 20, time.Minute)
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

func (m *GPTManager) requestHandler(clientId uint64) bool {
	return m.limiter.Allow(clientId)
}

func (m *GPTManager) MergeClient(id uint64, gptType enum.GPTType, historySize int) (IGPTClient, bool) {
	c, exists := m.clientsMap.merge(fmt.Sprintf(`%d`, id), func(s string) interface{} {
		c := CreateNewGPTClient(id, m.getApiKey(id), gptType, historySize, m.requestHandler)
		return c
	})
	return c.(IGPTClient), exists
}
