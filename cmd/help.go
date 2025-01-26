package cmd

import (
	"fmt"
)

func ShowHelp() {
	fmt.Println("pal - AI assistant for terminals")
	fmt.Println("Commands:")
	fmt.Println("  /help      Show this help message")
	fmt.Println("  /config    Configure API keys")
	fmt.Println("  /complete  Install shell completions")
	fmt.Println("  /ask       Ask general questions")
	fmt.Println("  /run       Run commands with AI assistance")
	fmt.Println()
	fmt.Println("First run 'pal /config' to set up your API key")
}
