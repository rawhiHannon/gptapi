package websockets

import (
	// "fmt"
	"github.com/google/uuid"
	// "strings"
	"log"
	"sync"
)

type Room struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	clients     map[*Client]bool
	clientsLock *sync.RWMutex
	liveSources map[string]int
}

func NewRoom(name string) *Room {
	return &Room{
		ID:          uuid.New(),
		Name:        name,
		clients:     make(map[*Client]bool),
		clientsLock: new(sync.RWMutex),
		liveSources: make(map[string]int),
	}
}

func (r *Room) registerClient(client *Client) {
	r.clientsLock.Lock()
	defer r.clientsLock.Unlock()
	r.clients[client] = true
	for key := range client.liveSources {
		if _, exist := r.liveSources[key]; !exist {
			r.liveSources[key] = 0
		}
		r.liveSources[key]++
	}
}

func (r *Room) unregisterClient(client *Client) {
	log.Println("Unregister client from room " + r.GetName())
	r.clientsLock.Lock()
	defer r.clientsLock.Unlock()
	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)
	}
	for key := range client.liveSources {
		if _, exist := r.liveSources[key]; exist {
			r.liveSources[key]--
		}
	}
	client.liveSources = make(map[string]bool)
}

func (r *Room) broadcastToAll(key string, message *Message) {
	r.clientsLock.RLock()
	defer r.clientsLock.RUnlock()
	for client := range r.clients {
		if key == "" || client.hasSource(key) {
			client.send <- message.encode()
		}
	}
}

func (r *Room) HasListeners(key string) bool {
	r.clientsLock.RLock()
	defer r.clientsLock.RUnlock()
	if len(r.clients) == 0 || r.clients == nil {
		return false
	}

	count, exist := r.liveSources[key]
	if key != "" && (!exist || count == 0) {
		return false
	}
	return true
}

func (r *Room) GetId() string {
	return r.ID.String()
}

func (r *Room) GetName() string {
	return r.Name
}
