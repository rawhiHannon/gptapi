package models

type Prompt struct {
	Text string
}

func NewPrompt(text string) Prompt {
	p := Prompt{
		Text: text,
	}
	return p
}
