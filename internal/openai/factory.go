package openai

import (
	"context"
	"gptapi/pkg/enum"
	"gptapi/pkg/models"
	"time"
)

func CreateNewGPTClient(id uint64, apiKey string, gptType enum.GPTType, window, limit int, rate time.Duration) models.IGPTClient {
	if apiKey == "" {
		return nil
	}
	switch gptType {
	case enum.GPT_3:
		return NewGPTClient(context.Background(), apiKey, nil)
	case enum.GPT_3_5_TURBO:
		return NewCGPTClient(id, apiKey, window, limit, rate)
	}
	return nil
}
