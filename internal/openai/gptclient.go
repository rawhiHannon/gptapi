package openai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
)

type GPTClient struct {
	apiKey      string
	client      gpt3.Client
	engine      string
	maxTokens   int
	temperature float32
	stream      func(string)
	prompt      string
	history     []string
	ctx         context.Context
}

func NewGPTClient(ctx context.Context, apiKey string, stream func(string)) *GPTClient {
	g := &GPTClient{
		stream:  stream,
		history: make([]string, 0),
		ctx:     ctx,
	}
	g.init(apiKey)
	return g
}

func (g *GPTClient) init(apiKey string) {
	g.apiKey = apiKey
	g.client = gpt3.NewClient(apiKey)
	g.engine = gpt3.TextDavinci003Engine
	g.maxTokens = 3000
	g.temperature = 0
}

func (g *GPTClient) SetRateLimitMsg(msg string) {

}

func (g *GPTClient) SetPrompt(prompt string, history []string) {
	g.prompt = prompt
	if history != nil {
		g.history = history
	} else {
		g.history = make([]string, 0)
	}
}

func (g *GPTClient) SendText(text string) (response string) {
	sb := strings.Builder{}
	err := g.client.CompletionStreamWithEngine(
		g.ctx,
		g.engine,
		gpt3.CompletionRequest{
			Prompt: []string{
				fmt.Sprintf(`%s\n%s\n%s`, g.prompt, strings.Join(g.history, "\n"), text),
			},
			MaxTokens:   gpt3.IntPtr(g.maxTokens),
			Temperature: gpt3.Float32Ptr(g.temperature),
		},
		func(resp *gpt3.CompletionResponse) {
			text := resp.Choices[0].Text
			if g.stream != nil {
				if text == "\n" && sb.Len() == 0 {
					return
				}
				g.stream(text)
			}
			sb.WriteString(text)
		},
	)
	if err != nil {
		log.Println(err)
		return ""
	}
	response = sb.String()
	response = strings.TrimLeft(response, "\n")
	g.history = append(g.history, text)
	g.history = append(g.history, response)
	return response
}
