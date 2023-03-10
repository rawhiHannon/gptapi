package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("gpt-cli: No commands provided")
		os.Exit(1)
	}

	switch os.Args[1] {
	default:
		fmt.Println(fmt.Sprintf(`gpt-cli: %s command not recognized`, os.Args[1]))
		os.Exit(1)
	}
}
