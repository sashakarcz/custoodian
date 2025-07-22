package cmd

import (
	"os"
	"path/filepath"
)

// readFile reads the entire content of a file
func readFile(filename string) ([]byte, error) {
	// Clean the file path to prevent directory traversal
	cleanPath := filepath.Clean(filename)
	
	return os.ReadFile(cleanPath)
}

// writeFile writes content to a file, creating directories as needed
func writeFile(filename, content string) error {
	// Clean the file path to prevent directory traversal
	cleanPath := filepath.Clean(filename)
	dir := filepath.Dir(cleanPath)
	
	// Use more restrictive directory permissions (0750)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// Use more restrictive file permissions (0600)
	return os.WriteFile(cleanPath, []byte(content), 0600)
}