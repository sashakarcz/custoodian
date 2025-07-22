package generator

import (
	"testing"

	"custoodian/pkg/config"
)

func TestNew(t *testing.T) {
	// Test creating generator with builtin templates
	gen, err := New("builtin")
	if err != nil {
		t.Errorf("Expected no error creating builtin generator, got: %v", err)
	}
	if gen == nil {
		t.Error("Expected generator to be created, got nil")
	}
}

func TestGenerate(t *testing.T) {
	// Create generator
	gen, err := New("builtin")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test generating with minimal config
	cfg := &config.Config{
		Project: &config.Project{
			Id:   "test-project-123",
			Name: "Test Project",
		},
	}

	files, err := gen.Generate(cfg)
	if err != nil {
		t.Errorf("Expected no error generating, got: %v", err)
	}

	// Check that some files were generated
	if len(files) == 0 {
		t.Error("Expected files to be generated, got none")
	}

	// Check that project.tf exists
	if _, exists := files["project.tf"]; !exists {
		t.Error("Expected project.tf to be generated")
	}

	// Check that variables.tf exists
	if _, exists := files["variables.tf"]; !exists {
		t.Error("Expected variables.tf to be generated")
	}
}