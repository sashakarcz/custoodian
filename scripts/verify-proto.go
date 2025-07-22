//go:build ignore

package main

import (
	"fmt"
	"os"
)

func main() {
	files := []string{
		"pkg/config/config.pb.go",
		"pkg/config/enums.pb.go",
	}

	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			fmt.Printf("ERROR: %s not found\n", file)
			os.Exit(1)
		}
		fmt.Printf("✓ %s exists\n", file)
	}
	
	fmt.Println("✓ All protobuf files generated successfully")
}