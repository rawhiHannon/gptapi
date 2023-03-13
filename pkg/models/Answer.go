package models

import "gptapi/pkg/enum"

type Answer struct {
	Data       string
	AnswerType enum.AnswerType
}

func NewAnswer(data string, answerType enum.AnswerType) *Answer {
	a := &Answer{
		Data:       data,
		AnswerType: answerType,
	}
	return a
}
