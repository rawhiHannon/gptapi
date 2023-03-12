package enum

type GPTType int

const (
	GPT_3 GPTType = iota
	GPT_3_5_TURBO
)

type SMTPType int

const (
	TLS SMTPType = iota
	SSL
)

type AnswerType int

const (
	TEXT_ANSWER AnswerType = iota
	IMAGE_ANSWER
)
