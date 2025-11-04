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
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput, err := inout.ReadStdin()
		if err != nil {
			return fmt.Errorf("error reading stdin: %v", err)
		}

		if stdinInput == "" {
			yoloMode, _ := cmd.Flags().GetBool("yolo")

			lastEditFilePath, pathErr := getLastEditOutputFilePath()
			if pathErr != nil {
				return fmt.Errorf("error getting last edit output file path: %v", pathErr)
			}

			lastEditContent, readErr := os.ReadFile(lastEditFilePath)
			if os.IsNotExist(readErr) {
				return fmt.Errorf("no input detected. Use with the output of the /edit command, or run /edit first")
			} else if readErr != nil {
				return fmt.Errorf("error reading previous edit output from %s: %v", lastEditFilePath, readErr)
			}

			fmt.Printf("No stdin input. Found previous edit output in %s.\n", lastEditFilePath)
			if !yoloMode {
				fmt.Print("Do you want to apply these edits? (y/N): ")
				var confirmation string
				_, scanErr := fmt.Scanln(&confirmation)
				if scanErr != nil {
					fmt.Fprint(os.Stderr, "Failed to read confirmation, assuming 'no'. Edit application cancelled.\n")
					return nil
				}
				if strings.ToLower(strings.TrimSpace(confirmation)) != "y" {
					fmt.Fprint(os.Stderr, "Edit application cancelled.\n")
					return nil
				}
			}
			stdinInput = string(lastEditContent)
		}

		edits, err := parseEdits(stdinInput)
		if err != nil {
			return fmt.Errorf("error parsing edits: %v", err)
		}

		if len(edits) == 0 {
			return fmt.Errorf("no valid edits found in input")
		}

		cfg := config.LoadConfigOrExit()

		// Get model for apply command
		applyModel := config.GetSelectedModel(cfg, "apply")
		client, err := ai.NewClient(cfg, applyModel)
		if err != nil {
			return fmt.Errorf("error creating AI client: %v", err)
		}

		appliedCount := 0
		for _, edit := range edits {
			err := applyEdit(client, edit, applyModel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying edit to %s: %v\n", edit.FilePath, err)
				continue
			}
			appliedCount++
			fmt.Printf("Applied edit to %s\n", edit.FilePath)
		}

		fmt.Printf("Successfully applied %d edit(s)\n", appliedCount)
		return nil
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

func applyEdit(client *ai.Client, edit Edit, model string) error {
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
	response, err := client.GetCompletion(context.Background(), applySystemPrompt, applyPrompt, false, 0.0, false, model)
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
