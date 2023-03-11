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
