package validator

import (
	"testing"

	"custoodian/pkg/config"
)

func TestValidateConfig(t *testing.T) {
	// Test empty config
	cfg := &config.Config{}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for empty config, got nil")
	}

	// Test config with project
	cfg = &config.Config{
		Project: &config.Project{
			Id:             "test-project-123",
			Name:           "Test Project",
			BillingAccount: "123456-ABCDEF-GHIJKL",
		},
	}
	err = ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestValidateProject(t *testing.T) {
	// Test nil project
	err := validateProject(nil)
	if err == nil {
		t.Error("Expected error for nil project, got nil")
	}

	// Test invalid project ID
	project := &config.Project{
		Id:   "invalid-project-id-that-is-way-too-long-for-gcp",
		Name: "Test",
	}
	err = validateProject(project)
	if err == nil {
		t.Error("Expected error for invalid project ID, got nil")
	}

	// Test valid project
	project = &config.Project{
		Id:   "test-project-123",
		Name: "Test Project",
	}
	err = validateProject(project)
	if err != nil {
		t.Errorf("Expected no error for valid project, got: %v", err)
	}
}

func TestIsValidGCPProjectID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"test-project-123", true},
		{"my-app-prod", true},
		{"short", false},                                           // too short
		{"invalid-project-id-that-is-way-too-long", false},       // too long
		{"Test-Project", false},                                   // uppercase
		{"test_project", false},                                   // underscore
		{"123-project", false},                                    // starts with number
		{"project-", false},                                       // ends with dash
	}

	for _, test := range tests {
		result := isValidGCPProjectID(test.id)
		if result != test.valid {
			t.Errorf("isValidGCPProjectID(%q) = %v, want %v", test.id, result, test.valid)
		}
	}
}