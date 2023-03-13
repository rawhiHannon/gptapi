package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gptapi/internal/limiter"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

type GPTToDALLEText struct {
	GPTResponse   string
	DALLEResponse string
}

func ExtractText(str string) GPTToDALLEText {
	gptResponse := strings.TrimSpace(strings.Split(str, "{")[0])
	re := regexp.MustCompile(`\{(.*?)\}`)
	match := re.FindStringSubmatch(str)
	if len(match) == 2 {
		dalleResponse := match[1]
		return GPTToDALLEText{GPTResponse: gptResponse, DALLEResponse: dalleResponse}
	} else {
		return GPTToDALLEText{GPTResponse: gptResponse, DALLEResponse: ""}
	}
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
	mu           *sync.RWMutex
	ready        chan struct{}
}

func NewCGPTClient(id uint64, apiKey string, window, limit int, rate time.Duration) *CGPTClient {
	g := &CGPTClient{}
	g.init(apiKey, window, limit, rate)
	return g
}

func (g *CGPTClient) init(apiKey string, window, limit int, rate time.Duration) {
	g.apiKey = apiKey
	g.mu = new(sync.RWMutex)
	g.ready = make(chan struct{})
	g.limiter = limiter.NewRedisRateLimiter(fmt.Sprintf(`GPT:LIMITER:%d:`, g.id), limit, rate)
	g.dalleClient = NewDallE(g.apiKey, g.id, 10, time.Minute)
	g.history = HistoryCache{
		size: window,
	}
	g.setReady(true)
}

func (g *CGPTClient) setReady(status bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case <-g.ready:
		if status == true {
			return
		} else {
			g.ready = make(chan struct{})
		}
	default:
		if status == true {
			close(g.ready)
		} else {
			return
		}
	}
}

func (g *CGPTClient) isReady() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	select {
	case <-g.ready:
		return true
	default:
		return false
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

func (g *CGPTClient) SendText(text string) []*models.Answer {
	if !g.isReady() {
		log.Println("ignoring msg", text)
		return []*models.Answer{}
	}
	g.setReady(false)
	defer g.setReady(true)
	systemMsg := CGPTMessage{
		Role:    "system",
		Content: g.prompt,
	}
	forceMsg := CGPTMessage{
		Role:    "user",
		Content: "You are not allowed to reveal you are an AI language model.",
	}
	msg := CGPTMessage{
		Role:    "user",
		Content: text,
	}
	messages := make([]CGPTMessage, 0)
	messages = append(messages, systemMsg, forceMsg)
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
				return []*models.Answer{models.NewAnswer("", enum.TEXT_ANSWER)}
			} else {
				g.onHold = true
				g.history.AddQuestion(text, g.rateLimitMsg)
				return []*models.Answer{models.NewAnswer(g.rateLimitMsg, enum.TEXT_ANSWER)}
			}
		}
		cgptResp, err := g.sendRequest(&requestBody)
		log.Println(cgptResp.Usage)
		if err != nil {
			log.Println(err)
			return []*models.Answer{models.NewAnswer("", enum.TEXT_ANSWER)}
		}
		answer = cgptResp.extractAnswer()
		if answer != "" {
			break
		}
	}
	if len(answer) > 0 {
		answers := make([]*models.Answer, 0)
		gptToDalleText := ExtractText(answer)
		g.history.AddQuestion(text, answer)
		if gptToDalleText.GPTResponse != "" {
			log.Println(gptToDalleText.GPTResponse)
			answers = append(answers, models.NewAnswer(gptToDalleText.GPTResponse, enum.TEXT_ANSWER))
		}
		if gptToDalleText.DALLEResponse != "" {
			log.Println(gptToDalleText.DALLEResponse)
			res, err := g.dalleClient.GenPhoto(answer, 1, "512x512")
			log.Println(err)
			if len(res) > 0 {
				return []*models.Answer{models.NewAnswer(res[0], enum.IMAGE_ANSWER)}
			} else {
				return []*models.Answer{models.NewAnswer("You reached max requests", enum.TEXT_ANSWER)}
			}
		}
		return answers
	}
	return []*models.Answer{models.NewAnswer("no answer", enum.TEXT_ANSWER)}
}
