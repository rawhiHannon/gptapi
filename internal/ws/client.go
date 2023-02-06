package websockets

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Client struct {
	conn        *websocket.Conn
	wsServer    *WsServer
	send        chan []byte
	ID          uuid.UUID `json:"id"`
	rooms       map[*Room]bool
	liveSources map[string]bool
}

func NewClient(conn *websocket.Conn, wsServer *WsServer) *Client {
	return &Client{
		ID:          uuid.New(),
		conn:        conn,
		wsServer:    wsServer,
		send:        make(chan []byte, 256),
		rooms:       make(map[*Room]bool),
		liveSources: make(map[string]bool),
	}

}

func (client *Client) readPump() {
	defer func() {
		log.Println("Closing client connection and channels", client.ID)
		client.disconnect()
	}()
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("unexpected close error", err)
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for room := range client.rooms {
		room.unregisterClient(client)
	}
	close(client.send)
	client.conn.Close()
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	log.Println(string(jsonMessage))
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}
	message.Sender = client
	switch message.Action {
	case SendStreamAction:
		roomID := message.Target.GetId()
		if room := client.wsServer.findRoomByID(roomID); room != nil {
			room.broadcastToAll("", &message)
		}
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	}
}

func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message
	data := message.Data
	if data != "" {
		liveSourcesArr := strings.Split(data, ",")
		for i := 0; i < len(liveSourcesArr); i++ {
			client.liveSources[liveSourcesArr[i]] = true
		}
	}
	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByName(message.Message)
	if room == nil {
		return
	}
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}
	room.unregisterClient(client)
}

func (client *Client) joinRoom(roomName string, sender *Client) {
	room := client.wsServer.findRoomByName(roomName)
	if room == nil {
		log.Println("no rooms with the name " + roomName)
		return
	}
	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.registerClient(client)
	}
}

func (client *Client) hasSource(key string) bool {
	if _, ok := client.liveSources[key]; ok {
		return true
	}
	return false
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}
