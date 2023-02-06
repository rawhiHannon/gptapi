package models

import (
	"fmt"
	"strings"
)

type Prompt struct {
	Text   string
	Params map[string]string
}

func NewPrompt(text string) Prompt {
	p := Prompt{
		Text:   text,
		Params: make(map[string]string),
	}
	return p
}

func (p *Prompt) Parse() string {
	parsedText := strings.Clone(p.Text)
	for k, v := range p.Params {
		key := fmt.Sprintf(`{{!%s}}`, k)
		parsedText = strings.ReplaceAll(parsedText, key, v)
	}
	return parsedText
}
