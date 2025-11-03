package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(selectCmd)
}

var selectCmd = &cobra.Command{
	Use:   "/select",
	Short: "Generate an HTML file to select files and folders",
	Long:  `Generates a static HTML page containing a file tree of the current directory. The page allows for selecting files and folders, and the output is a text list of the selected paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the current working directory.
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}

		// Create a temporary HTML file.
		tmpfile, err := ioutil.TempFile("", "pal-select-*.html")
		if err != nil {
			return fmt.Errorf("error creating temporary file: %w", err)
		}
		defer tmpfile.Close()

		// Generate the HTML content.
		html, err := generateHTML(dir)
		if err != nil {
			return fmt.Errorf("error generating HTML: %w", err)
		}

		// Write the HTML content to the temporary file.
		if _, err := tmpfile.WriteString(html); err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}

		// Print the path to the temporary file.
		fmt.Printf("File selector is available at: file://%s\n", tmpfile.Name())

		return nil
	},
}

//go:embed select.html
var selectHTML string

func generateHTML(dir string) (string, error) {
	// Get the file tree.
	fileTree, err := getFileTree(dir)
	if err != nil {
		return "", err
	}

	// Create a new template.
	tmpl, err := template.New("filetree").Parse(selectHTML)
	if err != nil {
		return "", err
	}

	// Execute the template with the file tree data.
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, fileTree)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

type File struct {
	Name     string
	Path     string
	Children []*File
}

func getFileTree(dir string) (*File, error) {
	root := &File{
		Name: filepath.Base(dir),
		Path: dir,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory.
		if path == dir {
			return nil
		}

		// Create a new file node.
		file := &File{
			Name: info.Name(),
			Path: path,
		}

		// Add the file to the tree.
		parent := findParent(root, filepath.Dir(path))
		if parent != nil {
			parent.Children = append(parent.Children, file)
		}

		return nil
	})

	return root, err
}

func findParent(root *File, path string) *File {
	if root.Path == path {
		return root
	}

	for _, child := range root.Children {
		if parent := findParent(child, path); parent != nil {
			return parent
		}
	}

	return nil
}


