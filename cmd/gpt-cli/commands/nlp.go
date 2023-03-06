package commands

import (
	"context"
	"fmt"
	"gptapi/internal/openai"
	"gptapi/pkg/utils"
	"os"
)

type NLPCommand struct {
	*Command
	Text *string
}

func CreateNLPCommand(args []string) *NLPCommand {
	s := &NLPCommand{
		Command: CreateCommand("nlp"),
	}
	s.Text = s.Cmd.String("d", "", "any text")
	if args != nil {
		s.Parse(args)
	}
	return s
}

func (s *NLPCommand) GetText() string {
	if s.Text != nil {
		return *s.Text
	}
	return ""
}

func (s *NLPCommand) ValidateArgs() {
	if s.IsParsed() == false {
		return
	}
	if s.GetText() == "" {
		s.Cmd.PrintDefaults()
		os.Exit(1)
	}
}

func (s *NLPCommand) Run() {
	utils.LoadEnv("")
	s.ValidateArgs()
	apiKey, ok := os.LookupEnv("GPT_API_KEY")
	if !ok || len(apiKey) == 0 {
		fmt.Println("No API Key provided")
	}
	c := openai.NewGPTClient(context.Background(), apiKey, func(s string) { fmt.Print(s) })
	_, err := c.SendText(s.GetText())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println()
	}
}
