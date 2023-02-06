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
