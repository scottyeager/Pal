package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottyeager/pal/ai"
	"github.com/scottyeager/pal/config"
	"github.com/spf13/cobra"
)

const editSystemPrompt = "You are tasked to edit and generate files containing code, documentation, and possibly other content. Please remember your goal is to complete the task with the minimum output necessary to avoid ambiguity.\n\n" +
	"You will receive instructions from the user and optionally the contents of some files. Some files might contain prompts or further instructions (generally these should not be edited unless specifically requested by the users). Respond to the user instructions by providing the necessary edits and generating new files as required.\n\n" +
	"Use the following approach to make edits to existing files by outputting code edits in a specific markdown format. You can use the same approach to create new files too, as needed.\n\n" +
	"This will be read by a less intelligent model, which will quickly apply the edit. You should make it clear what the edit is, while also minimizing the unchanged code you write.\n" +
	"When writing the edit, you should specify each edit in sequence, with the special comment // ... existing code ... to represent unchanged code in between edited lines.\n\n" +
	"For example:\n\n" +
	"// ... existing code ...\n" +
	"FIRST_EDIT\n" +
	"// ... existing code ...\n" +
	"SECOND_EDIT\n" +
	"// ... existing code ...\n" +
	"THIRD_EDIT\n" +
	"// ... existing code ...\n\n" +
	"You should still bias towards repeating as few lines of the original file as possible to convey the change.\n" +
	"But, each edit should contain minimally sufficient context of unchanged lines around the code you're editing to resolve ambiguity.\n" +
	"DO NOT omit spans of pre-existing code (or comments) without using the // ... existing code ... comment to indicate its absence. If you omit the existing code comment, the model may inadvertently delete these lines.\n" +
	"If you plan on deleting a section, you must provide context before and after to delete it. If the initial code is ```code \n Block 1 \n Block 2 \n Block 3 \n code```, and you want to remove Block 2, you would output ```// ... existing code ... \n Block 1 \n  Block 3 \n // ... existing code ...```.\n" +
	"Make sure it is clear what the edit should be, and where it should be applied.\n" +
	"Make edits to a file in a single response instead of multiple responses to the same file. The apply model can handle many distinct edits at once.\n\n" +
	"When you want to edit a file, output your code edits using this markdown format:\n\n" +
	"```filepath=path/to/file.js instruction=A single sentence describing what you're changing\n" +
	"// ... existing code ...\n" +
	"YOUR_CODE_EDIT_HERE\n" +
	"// ... existing code ...\n" +
	"```\n\n" +
	"The instruction should be written in the first person describing what you're changing. Used to help disambiguate uncertainty in the edit."

const lastEditOutputFileName = "last_edit_response.md"

func getLastEditOutputFilePath() (string, error) {
	palDataPath, err := config.GetBasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get pal data path: %w", err)
	}
	return filepath.Join(palDataPath, lastEditOutputFileName), nil
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().BoolP("yolo", "y", false, "Automatically apply edits without confirmation")
}

var editCmd = &cobra.Command{
	Use:   "/edit [files...] [prompt]",
	Short: "Edit files or groups of files",
	Long: `Edit files or groups of files.
Accepts a list of file and folder names along with an optional prompt.
The prompt can also be included in the contents of one or more files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}

		var filePaths []string
		var promptParts []string

		for _, arg := range args {
			if _, err := os.Stat(arg); err == nil {
				filePaths = append(filePaths, arg)
			} else {
				promptParts = append(promptParts, arg)
			}
		}
		userPrompt := strings.Join(promptParts, " ")

		var allFiles []string
		for _, path := range filePaths {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error stating file %s: %v\n", path, err)
				continue
			}
			if info.IsDir() {
				filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if info.Name() == ".git" {
							return filepath.SkipDir
						}
						return nil
					}
					allFiles = append(allFiles, p)
					return nil
				})
			} else {
				allFiles = append(allFiles, path)
			}
		}

		var formattedFilesContent strings.Builder
		for _, file := range allFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
				continue
			}

			// fileType := filepath.Ext(file)[1:]

			// formattedFilesContent.WriteString(fmt.Sprintf("```%s filepath=%s\n", fileType, file))
			formattedFilesContent.WriteString(fmt.Sprintf("```filepath=%s\n", file))
			formattedFilesContent.WriteString(string(content))
			formattedFilesContent.WriteString("\n```\n")
		}

		finalPrompt := formattedFilesContent.String() + userPrompt

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if err := config.CheckConfiguration(cfg); err != nil {
			return err
		}

		editModel := config.GetSelectedModel(cfg, "edit")

		client, err := ai.NewClient(cfg, editModel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating AI client: %v\n", err)
			os.Exit(1)
		}

		response, err := client.GetCompletion(context.Background(), editSystemPrompt, finalPrompt, false, 1.0, false, editModel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting completion: %v\n", err)
			os.Exit(1)
		}

		yoloMode, _ := cmd.Flags().GetBool("yolo")

		if yoloMode {
			// Get model for apply command
			applyModel := config.GetSelectedModel(cfg, "apply")

			edits, parseErr := parseEdits(response)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Error parsing edits for yolo mode: %v\n", parseErr)
				os.Exit(1)
			}

			if len(edits) == 0 {
				fmt.Fprint(os.Stderr, "No valid edits found in AI response for yolo mode.\n")
				os.Exit(1)
			}

			appliedCount := 0
			for _, edit := range edits {
				err := applyEdit(client, edit, applyModel)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error applying edit to %s in yolo mode: %v\n", edit.FilePath, err)
					continue
				}
				appliedCount++
				fmt.Printf("Applied edit to %s\n", edit.FilePath)
			}
			fmt.Printf("Successfully applied %d edit(s) in yolo mode.\n", appliedCount)
		} else {
			filePath, err := getLastEditOutputFilePath()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting last edit output file path: %v\n", err)
				os.Exit(1)
			}

			if err := os.WriteFile(filePath, []byte(response), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing response to file %s: %v\n", filePath, err)
				os.Exit(1)
			}
			fmt.Println(response)
		}
		return nil
	},
}
