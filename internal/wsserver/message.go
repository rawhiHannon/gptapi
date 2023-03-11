package wsserver

import (
	"encoding/json"
	"log"
)

const SendStreamAction = "stream"
const SendEventAction = "event"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const ChatAction = "chat"
const SettingsAction = "settings"

type Message struct {
	Id      string  `json:"id"`
	Action  string  `json:"action"`
	Message string  `json:"message"`
	Data    string  `json:"data"`
	Target  *Room   `json:"target"`
	Sender  *Client `json:"sender"`
}

func NewChatMessage(message string) *Message {
	messageObj := &Message{
		Action:  ChatAction,
		Message: message,
		Target:  nil,
		Sender:  nil,
	}
	return messageObj
}

func NewStreamMessage(room *Room, message string) *Message {
	messageObj := &Message{
		Action:  SendStreamAction,
		Message: message,
		Target:  room,
		Sender:  nil,
	}
	return messageObj
}

func NewEventMessage(room *Room, message string) *Message {
	messageObj := &Message{
		Action:  SendEventAction,
		Message: message,
		Target:  room,
		Sender:  nil,
	}
	return messageObj
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
