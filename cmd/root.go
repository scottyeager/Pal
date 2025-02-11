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

var fish bool
var fishAbbr bool
var zshAbbr bool
var fishCompletion bool
var zshCompletion bool
var temperature float64

var userMessage []string

func init() {
	rootCmd.Flags().BoolVarP(&fish, "version", "V", false, "Print version number")
	rootCmd.Flags().BoolVar(&fish, "fish", false, "Print fish abbreviation script and completion script, then exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "fish-abbr", false, "Print fish abbreviation script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshAbbr, "zsh-abbr", false, "Print zsh abbreviation script and exit. Output is meant to be sourced by zsh.")
	rootCmd.Flags().BoolVar(&fishCompletion, "fish-completion", false, "Print fish autocompletion script and exit. Output is meant to be sourced by fish.")
	rootCmd.Flags().BoolVar(&zshCompletion, "zsh-completion", false, "Print zsh autocompletion script and exit. Output is meant to be sourced by zsh.")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0, "Set the temperature for the AI model, between 0 and 2 (higher values make output more random)")

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
		return 2
	}

	if strings.HasPrefix(args[1], "-") {
		for i, arg := range args[2:] {
			if strings.HasPrefix(arg, "/") {
				return i + 3
			}
		}
		return len(args)
	}

	return 1
}

func Execute() {
	rootCmd.Version = version

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
		case "--zsh-abbr":
			path, err := abbr.InstallZshAbbr()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Println(path)
				os.Exit(0)
			}
		case "--zsh-completion":
			rootCmd.GenZshCompletionNoDesc(os.Stdout)
			os.Exit(0)
		case "--help", "-h", "--version", "-V", "__complete", "__completeNoDesc":
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
