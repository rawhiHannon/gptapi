package models

import "gptapi/pkg/enum"

type CacheManager interface {
	HSet(string, string, interface{}) error
	HGet(string, string) (string, error)
}

type IGPTClient interface {
	SetPrompt(string, []string)
	SetRateLimitMsg(string)
	SendText(string) (string, enum.AnswerType)
}
