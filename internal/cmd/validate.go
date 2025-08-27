package cmd

import (
	"fmt"

	"custoodian/internal/validator"

	"github.com/spf13/cobra"
)

type validateOptions struct {
	configFile string
}

func newValidateCmd() *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate a Protocol Buffer configuration file",
		Long: `Validate a Protocol Buffer text configuration file for syntax and constraints.

This command checks:
- Protocol Buffer syntax
- Field validation rules
- GCP resource constraints
- Cross-field dependencies
- Naming conventions

Examples:
  custodian validate config.textproto
  custodian validate examples/simple.textproto`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.configFile = args[0]
			return runValidate(opts)
		},
	}

	return cmd
}

func runValidate(opts *validateOptions) error {
	// Load configuration
	cfg, err := loadConfig(opts.configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := validator.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("✓ Configuration is valid")
	return nil
}

func init() {
	rootCmd.AddCommand(newValidateCmd())
}
