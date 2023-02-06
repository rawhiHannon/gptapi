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

func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

func (this *WsServer) ServeWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, this)
	go client.writePump()
	go client.readPump()
	this.register <- client
}

func (server *WsServer) BroadcastStream(key string, roomName string, data string) {
	room := server.findRoomByName(roomName)
	if room == nil {
		log.Println("no room with the name " + roomName)
		return
	}
	message := NewStreamMessage(room, data)
	room.broadcastToAll(key, message)
}

func (server *WsServer) BroadcastEvent(roomName string, data string) {
	room := server.findRoomByName(roomName)
	if room == nil {
		log.Println("no room with the name " + roomName)
		return
	}
	message := NewEventMessage(room, data)
	room.broadcastToAll("", message)
}

func (server *WsServer) CreateRoom(name string) *Room {
	server.roomsLock.Lock()
	defer server.roomsLock.Unlock()
	if server.rooms[name] != nil {
		return server.rooms[name]
	}
	room := NewRoom(name)
	server.rooms[name] = room
	return room
}

func (server *WsServer) HasRoom(room string) bool {
	server.roomsLock.RLock()
	defer server.roomsLock.RUnlock()

	return (server.rooms != nil && server.rooms[room] != nil)
}

func (server *WsServer) RoomHasListeners(key string, roomName string) bool {
	room := server.findRoomByName(roomName)
	return (room != nil && room.HasListeners(key))
}

func (server *WsServer) registerClient(client *Client) {
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
	}
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message
	}
}

func (server *WsServer) findRoomByName(name string) *Room {
	server.roomsLock.RLock()
	defer server.roomsLock.RUnlock()
	foundRoom := server.rooms[name]
	return foundRoom
}

func (server *WsServer) findRoomByID(ID string) *Room {
	server.roomsLock.RLock()
	defer server.roomsLock.RUnlock()

	var foundRoom *Room
	for _, room := range server.rooms {
		if room.GetId() == ID {
			foundRoom = room
			break
		}
	}
	return foundRoom
}
