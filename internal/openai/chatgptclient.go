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

type ChatGPTClient struct {
	apiKey  string
	prompt  string
	history HistoryCache
	retries int
}

func NewChatGPTClient(apiKey string) *ChatGPTClient {
	g := &ChatGPTClient{}
	g.init(apiKey)
	return g
}

func (g *ChatGPTClient) init(apiKey string) {
	g.apiKey = apiKey
	g.retries = DEFAULT_MAX_RETRIES
	g.history = HistoryCache{
		size: 100,
	}
}

func (g *ChatGPTClient) appendToHistory(question, answer string) {

}

func (g *ChatGPTClient) SetPrompt(prompt string, history []string) {
	g.prompt = prompt
	if history != nil {
		g.history.reset()
	} else {
		g.history.reset()
	}
}

func (g *ChatGPTClient) SendText(text string) (string, error) {
	systemMsg := CGPTMessage{
		Role:    "system",
		Content: g.prompt,
	}
	msg := CGPTMessage{
		Role:    "user",
		Content: text,
	}
	messages := make([]CGPTMessage, 0)
	messages = append(messages, systemMsg)
	messages = append(messages, g.history.GetMessages()...)
	messages = append(messages, msg)
	requestBody := CGPTRequest{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
	}
	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(postData))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", g.apiKey))
	resp, e := client.Do(req)
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var cgptResp CGPTResponse
	err = json.Unmarshal(body, &cgptResp)
	if err != nil {
		return "", err
	}
	answer := ""
	if cgptResp.Choices != nil || len(cgptResp.Choices) != 0 {
		answer = cgptResp.Choices[0].Message.Content
		log.Println(cgptResp.Usage)
	}
	g.history.AddQuestion(text, answer)
	return answer, nil
}
