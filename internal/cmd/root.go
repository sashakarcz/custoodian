package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "custoodian",
	Short: "Generate Terraform code from Protocol Buffer configurations for GCP",
	Long: `Custoodian is a tool that generates Terraform code from Protocol Buffer text configurations
for Google Cloud Platform resources. It provides type-safe infrastructure configuration
with comprehensive validation and supports custom template systems.

Inspired by Google's internal "latchkey" tool, Custoodian leverages Protocol Buffers
for strong typing and validation, catching configuration errors before Terraform runs.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}