package openai

import (
	"context"
	"gptapi/pkg/enum"
)

func CreateNewGPTClient(id uint64, apiKey string, gptType enum.GPTType, historySize int, reqHandler func(uint64) bool) IGPTClient {
	if apiKey == "" {
		return nil
	}
	switch gptType {
	case enum.GPT_3:
		return NewGPTClient(context.Background(), apiKey, nil)
	case enum.GPT_3_5_TURBO:
		return NewCGPTClient(id, apiKey, historySize, reqHandler)
	}
	return nil
}
