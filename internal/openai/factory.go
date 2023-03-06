package openai

import (
	"context"
	"gptapi/pkg/enum"
)

func CreateNewGPTClient(apiKey string, gptType enum.GPTType, historySize int) IGPTClient {
	if apiKey == "" {
		return nil
	}
	switch gptType {
	case enum.GPT_3:
		return NewGPTClient(context.Background(), apiKey, nil)
	case enum.GPT_3_5_TURBO:
		return NewCGPTClient(apiKey, historySize)
	}
	return nil
}
