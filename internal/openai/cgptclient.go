package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const DEFAULT_MAX_RETRIES = 5

type CGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CGPTRequest struct {
	Model    string        `json:"model"`
	Messages []CGPTMessage `json:"messages"`
}

type CGPTChoices struct {
	Index        int         `json:"index"`
	FinishReason string      `json:"finish_reason"`
	Message      CGPTMessage `json:"message"`
}

type CGPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CGPTResponse struct {
	Id      string        `json:"id"`
	Created int64         `json:"created"`
	Choices []CGPTChoices `json:"choices"`
	Usage   CGPTUsage     `json:"usage"`
}

func (c *CGPTResponse) extractAnswer() string {
	answer := ""
	if c.Choices != nil || len(c.Choices) != 0 {
		answer = c.Choices[0].Message.Content
	}
	return answer
}

const MAX_RETRIES = 3

type CGPTClient struct {
	id            uint64
	apiKey        string
	prompt        string
	maxReachedMsg string
	history       HistoryCache
	reqHandler    func(uint64) bool
	retries       int
}

func NewCGPTClient(id uint64, apiKey string, historySize int, reqHandler func(uint64) bool) *CGPTClient {
	g := &CGPTClient{}
	g.init(apiKey, historySize, reqHandler)
	return g
}

func (g *CGPTClient) init(apiKey string, historySize int, reqHandler func(uint64) bool) {
	g.apiKey = apiKey
	g.retries = DEFAULT_MAX_RETRIES
	g.reqHandler = reqHandler
	g.history = HistoryCache{
		size: historySize,
	}
}

func (g *CGPTClient) sendRequest(requestBody *CGPTRequest) (*CGPTResponse, error) {
	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(postData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", g.apiKey))
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var cgptResp CGPTResponse
	err = json.Unmarshal(body, &cgptResp)
	if err != nil {
		return nil, err
	}
	return &cgptResp, nil
}

func (g *CGPTClient) SetPrompt(prompt string, history []string) {
	g.prompt = prompt
	if history != nil {
		g.history.reset()
	} else {
		g.history.reset()
	}
}

func (g *CGPTClient) SetMaxReachedMsg(msg string) {
	g.maxReachedMsg = msg
}

func (g *CGPTClient) SendText(text string) (string, error) {
	systemMsg := CGPTMessage{
		Role:    "system",
		Content: g.prompt,
	}
	systemMsg2 := CGPTMessage{
		Role: "user",
		Content: `
		Rules:
		. you have to use the same words/concepts the user use.
		. you ask one question at a time about one specific thing.
		. when you can use a phrase from Phrases you have to use it onw phrase every answer.
		`,
	}
	msg := CGPTMessage{
		Role:    "user",
		Content: text,
	}
	messages := make([]CGPTMessage, 0)
	messages = append(messages, systemMsg, systemMsg2)
	messages = append(messages, g.history.GetMessages()...)
	messages = append(messages, msg)
	requestBody := CGPTRequest{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}
	answer := ""
	for i := 0; i < MAX_RETRIES; i++ {
		if g.reqHandler(g.id) == false {
			return g.maxReachedMsg, nil
		}
		cgptResp, err := g.sendRequest(&requestBody)
		log.Println(cgptResp.Usage)
		if err != nil {
			return "", err
		}
		answer = cgptResp.extractAnswer()
		if answer != "" {
			break
		}
	}
	g.history.AddQuestion(text, answer)
	return answer, nil
}
