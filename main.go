package main

import (
	"fmt"
	"os"
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

import (
	"github.com/yourusername/pal/cmd"
)

func showHelp() {
	cmd.ShowHelp()
}
