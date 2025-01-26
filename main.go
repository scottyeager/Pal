package main

import (
	"fmt"
	"os"

	"github.com/scottyeager/pal/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pal <command>")
		fmt.Println("Try 'pal /help' for more information")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "/help":
		showHelp()
	case "/config":
		cmd.Configure()
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func showHelp() {
	cmd.ShowHelp()
}
