package wsserver

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
}

func NewRoom(name string) *Room {
	return &Room{
		ID:          uuid.New(),
		Name:        name,
		clients:     make(map[*Client]bool),
		clientsLock: new(sync.RWMutex),
	}
}

func (r *Room) registerClient(client *Client) {
	r.clientsLock.Lock()
	defer r.clientsLock.Unlock()
	r.clients[client] = true
}

func (r *Room) unregisterClient(client *Client) {
	log.Println("Unregister client from room " + r.GetName())
	r.clientsLock.Lock()
	defer r.clientsLock.Unlock()
	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)
	}
}

func (r *Room) broadcastToAll(key string, message *Message) {
	r.clientsLock.RLock()
	defer r.clientsLock.RUnlock()
	for client := range r.clients {
		client.send <- message.encode()
	}
}

func (r *Room) HasListeners(key string) bool {
	r.clientsLock.RLock()
	defer r.clientsLock.RUnlock()
	if len(r.clients) == 0 || r.clients == nil {
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
