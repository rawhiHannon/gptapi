package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gptapi/internal/limiter"
	"gptapi/pkg/enum"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

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

const (
	MAX_RETRIES = 3
	CHATGPT_API = "https://api.openai.com/v1/chat/completions"
	GPT_MODEL   = "gpt-3.5-turbo"
)

type CGPTClient struct {
	id           uint64
	apiKey       string
	prompt       string
	rateLimitMsg string
	onHold       bool
	history      HistoryCache
	dalleClient  *DallE
	limiter      *limiter.RedisRateLimiter
}

func NewCGPTClient(id uint64, apiKey string, window, limit int, rate time.Duration) *CGPTClient {
	g := &CGPTClient{}
	g.init(apiKey, window, limit, rate)
	return g
}

func (g *CGPTClient) init(apiKey string, window, limit int, rate time.Duration) {
	g.apiKey = apiKey
	g.limiter = limiter.NewRedisRateLimiter(fmt.Sprintf(`GPT:LIMITER:%d:`, g.id), limit, rate)
	g.dalleClient = NewDallE(g.apiKey, g.id, 10, time.Minute)
	g.history = HistoryCache{
		size: window,
	}
}

func (g *CGPTClient) sendRequest(requestBody *CGPTRequest) (*CGPTResponse, error) {
	postData, _ := json.Marshal(requestBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", CHATGPT_API, bytes.NewReader(postData))
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

func (g *CGPTClient) SetRateLimitMsg(msg string) {
	g.rateLimitMsg = msg
}

func (g *CGPTClient) SendText(text string) (string, enum.AnswerType) {
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
		Model:    GPT_MODEL,
		Messages: messages,
	}
	answer := ""
	for i := 0; i < MAX_RETRIES; i++ {
		if g.limiter.Allow(g.id) == false {
			if g.onHold == true {
				return "", enum.TEXT_ANSWER
			} else {
				g.onHold = true
				g.history.AddQuestion(text, g.rateLimitMsg)
				return g.rateLimitMsg, enum.TEXT_ANSWER
			}
		}
		cgptResp, err := g.sendRequest(&requestBody)
		log.Println(cgptResp.Usage)
		if err != nil {
			log.Println(err)
			return "", enum.TEXT_ANSWER
		}
		answer = cgptResp.extractAnswer()
		if answer != "" {
			break
		}
	}
	if len(answer) > 1 && strings.HasPrefix(answer, "{") {
		res, _ := g.dalleClient.GenPhoto(answer, 1, "512x512")
		log.Println(answer)
		return res[0], enum.IMAGE_ANSWER
	}
	g.history.AddQuestion(text, answer)
	return answer, enum.TEXT_ANSWER
}
