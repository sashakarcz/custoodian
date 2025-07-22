//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Create pkg/config directory if it doesn't exist
	err := os.MkdirAll("pkg/config", 0755)
	if err != nil {
		fmt.Printf("Error creating pkg/config directory: %v\n", err)
		os.Exit(1)
	}

	// Find all .pb.go files in proto/custoodian/
	protoDir := "proto/custoodian"
	files, err := filepath.Glob(filepath.Join(protoDir, "*.pb.go"))
	if err != nil {
		fmt.Printf("Error finding .pb.go files: %v\n", err)
		os.Exit(1)
	}

	// Move .pb.go files to pkg/config/
	for _, file := range files {
		basename := filepath.Base(file)
		dest := filepath.Join("pkg/config", basename)
		err := os.Rename(file, dest)
		if err != nil {
			fmt.Printf("Error moving %s to %s: %v\n", file, dest, err)
			os.Exit(1)
		}
		fmt.Printf("Moved %s to %s\n", file, dest)
	}

	// Remove .pb.validate.go files if they exist
	validateFiles, err := filepath.Glob(filepath.Join(protoDir, "*.pb.validate.go"))
	if err == nil {
		for _, file := range validateFiles {
			os.Remove(file)
			fmt.Printf("Removed %s\n", file)
		}
	}

	if len(files) == 0 {
		fmt.Println("No .pb.go files found to move")
	}
}