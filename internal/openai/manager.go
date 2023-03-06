package openai

import (
	"gptapi/internal/limiter"
	"gptapi/pkg/enum"
	"os"
	"strings"
	"sync"
	"time"
)

type IGPTClient interface {
	SetPrompt(string, []string)
	SendText(string) (string, error)
}

type GPTManager struct {
	clients map[string]IGPTClient
	apiKeys []string
	limiter *limiter.RedisRateLimiter
	mu      *sync.RWMutex
}

func NewGPTManager() *GPTManager {
	m := &GPTManager{
		clients: make(map[string]IGPTClient),
		mu:      new(sync.RWMutex),
	}
	m.limiter = limiter.NewRedisRateLimiter("GPT:LIMITER:", 10, time.Minute)
	m.loadApiKeys()
	return m
}

func (m *GPTManager) loadApiKeys() {
	keys, ok := os.LookupEnv("GPT_API_KEY")
	if !ok || len(keys) == 0 {
		panic("OpenAI: api keys missing")
	}
	m.apiKeys = strings.Split(keys, ",")
}

func (m *GPTManager) getApiKey() string {
	return m.apiKeys[0]
}

func (m *GPTManager) AddClient(id string, gptType enum.GPTType, stream func(string)) IGPTClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, exists := m.clients[id]; exists {
		return client
	}
	c := CreateNewGPTClient(m.getApiKey(), gptType, stream)
	m.clients[id] = c
	return c
}

func (m *GPTManager) GetClient(id string) IGPTClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.limiter.Allow(id) {
		return nil
	}
	if client, exists := m.clients[id]; exists {
		return client
	}
	return nil
}
