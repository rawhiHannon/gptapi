package gpt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gptapi/pkg/models"
	"gptapi/pkg/utils"

	"github.com/PullRequestInc/go-gpt3"
)

type GPTClient struct {
	client      gpt3.Client
	engine      string
	maxTokens   int
	temperature float32
	stream      func(string)
	prompt      models.Prompt
	history     []string
	ctx         context.Context
}

func NewGPTClient(ctx context.Context, stream func(string)) *GPTClient {
	g := &GPTClient{
		stream:  stream,
		prompt:  models.NewPrompt(""),
		history: make([]string, 0),
		ctx:     ctx,
	}
	g.init()
	return g
}

func (g *GPTClient) init() {
	utils.LoadEnv("")
	apiKey, ok := os.LookupEnv("GPT_API_KEY")
	if !ok {
		panic("Missing GPT_API_KEY")
	}
	g.client = gpt3.NewClient(apiKey)
	g.engine = gpt3.TextDavinci003Engine
	g.maxTokens = 3000
	g.temperature = 0
}

func (g *GPTClient) SetPrompt(prompt models.Prompt, history []string) {
	g.prompt = prompt
	if history != nil {
		g.history = history
	} else {
		g.history = make([]string, 0)
	}
}

func (g *GPTClient) SendText(text string) (response string, err error) {
	sb := strings.Builder{}
	err = g.client.CompletionStreamWithEngine(
		g.ctx,
		g.engine,
		gpt3.CompletionRequest{
			Prompt: []string{
				fmt.Sprintf(`%s\n%s\n%s`, g.prompt.Text, strings.Join(g.history, "\n"), text),
			},
			MaxTokens:   gpt3.IntPtr(g.maxTokens),
			Temperature: gpt3.Float32Ptr(g.temperature),
		},
		func(resp *gpt3.CompletionResponse) {
			text := resp.Choices[0].Text
			if g.stream != nil {
				if text == "\n" {
					return
				}
				g.stream(text)
			}
			sb.WriteString(text)
		},
	)
	if err != nil {
		return "", err
	}
	response = sb.String()
	response = strings.TrimLeft(response, "\n")
	g.history = append(g.history, text)
	g.history = append(g.history, response)
	return response, nil
}
