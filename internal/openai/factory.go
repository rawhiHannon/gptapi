package openai

import (
	"context"
	"gptapi/pkg/enum"
)

func CreateNewGPTClient(apiKey string, gptType enum.GPTType, stream func(string)) IGPTClient {
	switch gptType {
	case enum.GPT_3:
		return NewGPTClient(context.Background(), apiKey, stream)
	case enum.GPT_3_5_TURBO:
		return NewChatGPTClient(apiKey)
	}
	return nil
}
