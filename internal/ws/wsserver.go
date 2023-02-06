package websockets

import (
	"log"
	"net/http"
	"sync"
)

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[string]*Room
	roomsLock  *sync.RWMutex
}

func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[string]*Room),
		roomsLock:  new(sync.RWMutex),
	}
}

func (s *WsServer) registerClient(client *Client) {
	s.clients[client] = true
}

func (s *WsServer) unregisterClient(client *Client) {
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
	}
}

func (s *WsServer) broadcastToClients(message []byte) {
	for client := range s.clients {
		client.send <- message
	}
}

func (s *WsServer) findRoomByName(name string) *Room {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()
	foundRoom := s.rooms[name]
	return foundRoom
}

func (s *WsServer) findRoomByID(ID string) *Room {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()

	var foundRoom *Room
	for _, room := range s.rooms {
		if room.GetId() == ID {
			foundRoom = room
			break
		}
	}
	return foundRoom
}

func (s *WsServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.unregister:
			s.unregisterClient(client)
		case message := <-s.broadcast:
			s.broadcastToClients(message)
		}
	}
}

func (s *WsServer) ServeWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, s)
	go client.writePump()
	go client.readPump()
	s.register <- client
}

func (s *WsServer) BroadcastStream(key string, roomName string, data string) {
	room := s.findRoomByName(roomName)
	if room == nil {
		log.Println("no room with the name " + roomName)
		return
	}
	message := NewStreamMessage(room, data)
	room.broadcastToAll(key, message)
}

func (s *WsServer) BroadcastEvent(roomName string, data string) {
	room := s.findRoomByName(roomName)
	if room == nil {
		log.Println("no room with the name " + roomName)
		return
	}
	message := NewEventMessage(room, data)
	room.broadcastToAll("", message)
}

func (s *WsServer) CreateRoom(name string) *Room {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()
	if s.rooms[name] != nil {
		return s.rooms[name]
	}
	room := NewRoom(name)
	s.rooms[name] = room
	return room
}

func (s *WsServer) HasRoom(room string) bool {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()

	return (s.rooms != nil && s.rooms[room] != nil)
}

func (s *WsServer) RoomHasListeners(key string, roomName string) bool {
	room := s.findRoomByName(roomName)
	return (room != nil && room.HasListeners(key))
}
