package gpt

import (
	"context"
	"sync"
)

type GPTManager struct {
	clients map[string]*GPTClient
	mu      *sync.RWMutex
}

func NewGPTManager() *GPTManager {
	m := &GPTManager{
		clients: make(map[string]*GPTClient),
		mu:      new(sync.RWMutex),
	}
	return m
}

func (m *GPTManager) AddClient(id string, stream func(string)) *GPTClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, exists := m.clients[id]; exists {
		return client
	}
	c := NewGPTClient(context.Background(), stream)
	m.clients[id] = c
	return c
}

func (m *GPTManager) GetClient(id string) *GPTClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if client, exists := m.clients[id]; exists {
		return client
	}
	return nil
}
