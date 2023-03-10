package wsserver

import (
	"encoding/json"
	"gptapi/internal/uniqid"
	"log"
	"time"

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
	ID              uint64 `json:"id"`
	Token           string `json:"token"`
	conn            *websocket.Conn
	wsServer        *WsServer
	send            chan []byte
	messageHandler  func(*Client, *Message)
	settingsHandler func(*Client, *Message)
	rooms           map[*Room]bool
}

func NewClient(conn *websocket.Conn, wsServer *WsServer, token string) *Client {
	return &Client{
		ID:       uniqid.NextId(),
		Token:    token,
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
		rooms:    make(map[*Room]bool),
	}
}

func (c *Client) readPump() {
	defer func() {
		log.Println("Closing client connection and channels", c.ID)
		c.disconnect()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("unexpected close error", err)
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		c.handleNewMessage(jsonMessage)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) disconnect() {
	c.wsServer.unregister <- c
	for room := range c.rooms {
		room.unregisterClient(c)
	}
	close(c.send)
	c.conn.Close()
}

func (c *Client) handleNewMessage(jsonMessage []byte) {
	log.Println(string(jsonMessage))
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}
	if message.Action == SettingsAction {
		if c.settingsHandler != nil {
			c.settingsHandler(c, &message)
		}
	} else {
		if c.messageHandler != nil {
			c.messageHandler(c, &message)
		}
	}
	message.Sender = c
	switch message.Action {
	case SendStreamAction:
		roomID := message.Target.GetId()
		if room := c.wsServer.findRoomByID(roomID); room != nil {
			room.broadcastToAll(&message)
		}
	case JoinRoomAction:
		c.handleJoinRoomMessage(message)
	case ChatAction:
		c.handleChatMessage(message)
	case LeaveRoomAction:
		c.handleLeaveRoomMessage(message)
	}
}

func (c *Client) handleChatMessage(message Message) {
	// TODO:
}

func (c *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message
	c.joinRoom(roomName, nil)
}

func (c *Client) handleLeaveRoomMessage(message Message) {
	room := c.wsServer.findRoomByName(message.Message)
	if room == nil {
		return
	}
	if _, ok := c.rooms[room]; ok {
		delete(c.rooms, room)
	}
	room.unregisterClient(c)
}

func (c *Client) joinRoom(roomName string, sender *Client) {
	room := c.wsServer.findRoomByName(roomName)
	if room == nil {
		log.Println("no rooms with the name " + roomName)
		return
	}
	if !c.isInRoom(room) {
		c.rooms[room] = true
		room.registerClient(c)
	}
}

func (c *Client) isInRoom(room *Room) bool {
	if _, ok := c.rooms[room]; ok {
		return true
	}
	return false
}

func (s *Client) SetOnMessageReceived(handler func(*Client, *Message)) {
	s.messageHandler = handler
}

func (s *Client) SetOnSettingsReceived(handler func(*Client, *Message)) {
	s.settingsHandler = handler
}

func (c *Client) Send(message *Message) {
	c.send <- message.encode()
}
