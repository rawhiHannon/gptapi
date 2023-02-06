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

func (room *Room) registerClient(client *Client) {
	room.clientsLock.Lock()
	defer room.clientsLock.Unlock()
	room.clients[client] = true
	for key := range client.liveSources {
		if _, exist := room.liveSources[key]; !exist {
			room.liveSources[key] = 0
		}
		room.liveSources[key]++
	}
}

func (room *Room) unregisterClient(client *Client) {
	log.Println("Unregister client from room " + room.GetName())
	room.clientsLock.Lock()
	defer room.clientsLock.Unlock()
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}
	for key := range client.liveSources {
		if _, exist := room.liveSources[key]; exist {
			room.liveSources[key]--
		}
	}
	client.liveSources = make(map[string]bool)
}

func (room *Room) broadcastToAll(key string, message *Message) {
	room.clientsLock.RLock()
	defer room.clientsLock.RUnlock()
	for client := range room.clients {
		if key == "" || client.hasSource(key) {
			client.send <- message.encode()
		}
	}
}

func (room *Room) HasListeners(key string) bool {
	room.clientsLock.RLock()
	defer room.clientsLock.RUnlock()
	if len(room.clients) == 0 || room.clients == nil {
		return false
	}

	count, exist := room.liveSources[key]
	if key != "" && (!exist || count == 0) {
		return false
	}
	return true
}

func (room *Room) GetId() string {
	return room.ID.String()
}

func (room *Room) GetName() string {
	return room.Name
}
