package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/scottyeager/pal/abbr"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

var version string

var temperature float64
var markdown bool

var userMessage []string

func init() {
	rootCmd.Flags().Bool("version", false, "Print version number")
	rootCmd.Flags().Bool("fish", false, "Print fish abbreviation script and completion script, then exit. Output is meant to be sourced by fish")
	rootCmd.Flags().Bool("zsh", false, "Print zsh abbreviation script and completion script, then exit. Output is meant to be sourced by zsh")
	rootCmd.Flags().Bool("fish-config", false, "Outputs lines mean to be appended to fish.config, to enable autocompletions and abbreviations")
	rootCmd.Flags().Bool("zsh-config", false, "Outputs lines mean to be appended to ~/.zshrc, to enable autocompletions and abbreviations")
	rootCmd.Flags().Bool("fish-abbr", false, "Print fish abbreviation script and exit. Output is meant to be sourced by fish")
	rootCmd.Flags().Bool("zsh-abbr", false, "Writes the zsh-abbr plugin to a tmp directory and prints the path, to be sourced by Zsh")
	rootCmd.Flags().Bool("fish-completion", false, "Print fish autocompletion script and exit. Output is meant to be sourced by fish")
	rootCmd.Flags().Bool("zsh-completion", false, "Print zsh autocompletion script and exit. Output is meant to be sourced by zsh")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0, "Set the temperature for the AI model, between 0 and 2 (higher values make output more random)")
	rootCmd.PersistentFlags().BoolVarP(&markdown, "markdown", "m", false, "Toggle markdown formatting in output (inverts your config setting)")

	// Disable help command. --help still works
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

var rootCmd = &cobra.Command{
	// Use:   "pal", or whatever the user calls it
	Use:   os.Args[0],
	Short: "pal is a command-line tool that suggests shell commands",
	Long: `pal is a command-line tool that suggests shell commands based on your input.
It uses AI to generate commands and can also manage shell abbreviations.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if temperature < 0 || temperature > 2 {
			return fmt.Errorf("Temperature must be between 0 and 2")
		} else {
			return nil
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Commands(*cmd, args)
	},
}

func preparse(args []string) int {
	if strings.HasPrefix(args[1], "/") {
		// If a command takes user message, then everything after the command
		// is user message. Otherwise, pass any args/flags to Cobra for parsing
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == args[1] {
				if _, ok := cmd.Annotations["takes_user_message"]; !ok {
					return len(args)
				}
			}
		}
		return 2
	}

	// If first arg is a flag, look for a command after it
	if strings.HasPrefix(args[1], "-") {
		for i, arg := range args {
			if strings.HasPrefix(arg, "/") {
				return i + 1
			}
		}
	}

	// Default case - no command found, all args are user message
	return 1
}

func Execute() {
	if version != "" {
		rootCmd.Version = version
	} else {
		rootCmd.Version = "dev"
	}

	// We define these as Cobra flags, so that help and autocomplete works, but
	// we handle them straight out of the gate here and bypass Cobra. One reason
	// is to simplify the preparsing. Another reason is that these commands
	// produce output that is meant to be sourced (executed), and we want to be
	// certain that AI generated commands are never output instead

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--fish":
			cfg := config.LoadConfigOrExit()
			fmt.Println(abbr.GetFishAbbrScript(cfg.AbbreviationPrefix))
			rootCmd.GenFishCompletion(os.Stdout, true)
			// Disables file name completions. Set command name dynamically in
			// case the user changed it
			fmt.Printf("complete -c %s -f", os.Args[0])
			os.Exit(0)
		case "--fish-config":
			fmt.Println("\n# The following line enables autocompletions and abbreviations for pal")
			fmt.Println("# Just remove or comment the line to undo all changes to your shell")
			fmt.Printf("%s --fish | source\n", os.Args[0])
			os.Exit(0)
		case "--fish-abbr":
			cfg := config.LoadConfigOrExit()
			fmt.Println(abbr.GetFishAbbrScript(cfg.AbbreviationPrefix))
			os.Exit(0)
		case "--fish-completion":
			rootCmd.GenFishCompletion(os.Stdout, true)
			// Disables file name completions. Set command name dynamically in
			// case the user changed it
			fmt.Printf("complete -c %s -f", os.Args[0])
			os.Exit(0)
		case "--zsh-config":
			fmt.Println("\n# The following line enables autocompletions and abbreviations for pal")
			fmt.Println("# Just remove or comment the line to undo all changes to your shell")
			fmt.Printf("source <(%s --zsh)\n", os.Args[0])
			os.Exit(0)
		case "--zsh":
			cfg := config.LoadConfigOrExit()
			fmt.Println(abbr.GetZshAbbrScript(cfg.AbbreviationPrefix))
			rootCmd.GenZshCompletionNoDesc(os.Stdout)
			os.Exit(0)
		case "--zsh-abbr":
			cfg := config.LoadConfigOrExit()
			fmt.Println(abbr.GetZshAbbrScript(cfg.AbbreviationPrefix))
			os.Exit(0)
		case "--zsh-completion":
			rootCmd.GenZshCompletionNoDesc(os.Stdout)
			os.Exit(0)
		case "--help", "-h", "--version", "__complete", "__completeNoDesc":
			// No-op here, just skipping preparsing
		default:
			split := preparse(os.Args)
			userMessage = os.Args[split:]
			os.Args = os.Args[:split]
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
