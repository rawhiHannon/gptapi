package openai

type HistoryCache struct {
	messages []CGPTMessage
	size     int
}

func NewHistoryCache(windowSize int) *HistoryCache {
	return &HistoryCache{
		messages: make([]CGPTMessage, 0),
		size:     windowSize,
	}
}

func (h *HistoryCache) reset() {
	h.messages = make([]CGPTMessage, 0)
}

func (h *HistoryCache) AddMessage(msg CGPTMessage) {
	h.messages = append(h.messages, msg)
	if len(h.messages)%2 == 0 && len(h.messages) > h.size {
		h.messages = h.messages[1:]
	}
}

func (h *HistoryCache) AddQuestion(question, answer string) {
	h.AddMessage(CGPTMessage{
		Role:    "user",
		Content: question,
	})
	h.AddMessage(CGPTMessage{
		Role:    "assistant",
		Content: answer,
	})
}

func (h *HistoryCache) GetMessages() []CGPTMessage {
	return h.messages
}
