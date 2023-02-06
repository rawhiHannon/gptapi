package commands

import (
	"flag"
)

type Command struct {
	Cmd *flag.FlagSet
}

func CreateCommand(name string) *Command {
	s := &Command{}
	s.Cmd = flag.NewFlagSet(name, flag.ExitOnError)
	return s
}

func (s *Command) Parse(args []string) {
	s.Cmd.Parse(args)
}

func (s *Command) IsParsed() bool {
	return s.Cmd.Parsed()
}
