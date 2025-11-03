package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/scottyeager/pal/inout"
	"github.com/spf13/cobra"
)

const applySystemPrompt = "You are a code editing assistant that receives file edits and returns the complete, updated file content. Your task is to apply the provided edit instructions to the original code and return the full file with changes applied.\n\n" +
	"You will receive the original code, an instruction describing what to change, and the specific update to apply. Apply the changes precisely and return the complete updated file content.\n\n" +
	"Format your response as just the complete file content with all changes applied. Do not include any explanations, comments, or markdown formatting."

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolP("yolo", "y", false, "Automatically apply edits without confirmation when using previous edit output")
}

var applyCmd = &cobra.Command{
	Use:   "/apply",
	Short: "Apply edits to files",
	Long: `Apply edits to files.
Reads edit instructions from stdin and applies them to the specified files.
Use with the output of the /edit command.`,
	Run: func(cmd *cobra.Command, args []string) {
		stdinInput, err := inout.ReadStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}

		if stdinInput == "" {
			yoloMode, _ := cmd.Flags().GetBool("yolo")

			lastEditFilePath, pathErr := getLastEditOutputFilePath()
			if pathErr != nil {
				fmt.Fprintf(os.Stderr, "Error getting last edit output file path: %v\n", pathErr)
				os.Exit(1)
			}

			lastEditContent, readErr := os.ReadFile(lastEditFilePath)
			if os.IsNotExist(readErr) {
				fmt.Fprint(os.Stderr, "No input detected. Use with the output of the /edit command, or run /edit first.\n")
				os.Exit(1)
			} else if readErr != nil {
				fmt.Fprintf(os.Stderr, "Error reading previous edit output from %s: %v\n", lastEditFilePath, readErr)
				os.Exit(1)
			}

			fmt.Printf("No stdin input. Found previous edit output in %s.\n", lastEditFilePath)
			if !yoloMode {
				fmt.Print("Do you want to apply these edits? (y/N): ")
				var confirmation string
				_, scanErr := fmt.Scanln(&confirmation)
				if scanErr != nil {
					fmt.Fprint(os.Stderr, "Failed to read confirmation, assuming 'no'. Edit application cancelled.\n")
					os.Exit(0)
				}
				if strings.ToLower(strings.TrimSpace(confirmation)) != "y" {
					fmt.Fprint(os.Stderr, "Edit application cancelled.\n")
					os.Exit(0)
				}
			}
			stdinInput = string(lastEditContent)
		}

		edits, err := parseEdits(stdinInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing edits: %v\n", err)
			os.Exit(1)
		}

		if len(edits) == 0 {
			fmt.Fprint(os.Stderr, "No valid edits found in input.\n")
			os.Exit(1)
		}

		cfg := config.LoadConfigOrExit()
		client, err := ai.NewClient(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		appliedCount := 0
		for _, edit := range edits {
			err := applyEdit(client, edit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying edit to %s: %v\n", edit.FilePath, err)
				continue
			}
			appliedCount++
			fmt.Printf("Applied edit to %s\n", edit.FilePath)
		}

		fmt.Printf("Successfully applied %d edit(s)\n", appliedCount)
	},
}

type Edit struct {
	FilePath    string
	Instruction string
	Update      string
}

func parseEdits(input string) ([]Edit, error) {
	var edits []Edit
	lines := strings.Split(input, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for the start of a code block with filepath and instruction
		if strings.HasPrefix(line, "```filepath=") {
			// Parse filepath and instruction from the opening line
			openingLine := strings.TrimPrefix(line, "```filepath=")

			// Split on the first space to separate filepath from instruction
			parts := strings.SplitN(openingLine, " ", 2)
			if len(parts) < 1 {
				continue
			}

			filePath := strings.TrimSpace(parts[0])
			instruction := ""
			if len(parts) > 1 {
				instruction = strings.TrimSpace(parts[1])
			}

			// Collect content until we find the closing ```
			var codeLines []string
			i++ // Move to next line
			for i < len(lines) {
				line := strings.TrimSpace(lines[i])
				if line == "```" {
					break
				}
				codeLines = append(codeLines, lines[i])
				i++
			}

			codeContent := strings.TrimSpace(strings.Join(codeLines, "\n"))

			// Skip if no actual content
			if codeContent == "" {
				continue
			}

			edit := Edit{
				FilePath:    filePath,
				Instruction: instruction,
				Update:      codeContent,
			}

			edits = append(edits, edit)
		}
	}

	return edits, nil
}

func applyEdit(client *ai.Client, edit Edit) error {
	// Read the original file content
	originalContent, err := os.ReadFile(edit.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Format the Apply API request
	applyPrompt := fmt.Sprintf("<instruction>%s</instruction>\n<code>%s</code>\n<update>%s</update>",
		edit.Instruction,
		string(originalContent),
		edit.Update,
	)

	// Get the completion from the AI
	response, err := client.GetCompletion(context.Background(), applySystemPrompt, applyPrompt, false, 0.0, false)
	if err != nil {
		return fmt.Errorf("failed to get completion: %w", err)
	}

	// Write the response to the file
	err = os.WriteFile(edit.FilePath, []byte(response), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
