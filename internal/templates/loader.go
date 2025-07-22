package templates

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LoadFromDirectory loads templates from a local directory
func LoadFromDirectory(dir string) (map[string]string, error) {
	templates := make(map[string]string)
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-template files
		if info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}
		
		// Read template content
		content, err := readFileContent(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}
		
		// Use relative path as template name
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		
		templates[relPath] = content
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to load templates from directory %s: %w", dir, err)
	}
	
	if len(templates) == 0 {
		return nil, fmt.Errorf("no template files found in directory %s", dir)
	}
	
	return templates, nil
}

// LoadFromGit loads templates from a Git repository
func LoadFromGit(repoURL string) (map[string]string, error) {
	// TODO: Implement Git repository cloning and template loading
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("Git repository template loading not yet implemented")
}

// readFileContent reads the entire content of a file
func readFileContent(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}