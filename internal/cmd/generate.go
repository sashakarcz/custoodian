package cmd

import (
	"fmt"
	"path/filepath"

	"custodian/internal/generator"
	"custodian/internal/validator"
	"custodian/pkg/config"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/prototext"
)

type generateOptions struct {
	configFile   string
	outputDir    string
	templateDir  string
	templateRepo string
	validate     bool
	dryRun       bool
}

func newGenerateCmd() *cobra.Command {
	opts := &generateOptions{
		validate: true,
	}

	cmd := &cobra.Command{
		Use:   "generate [config-file]",
		Short: "Generate Terraform code from Protocol Buffer configuration",
		Long: `Generate Terraform code from a Protocol Buffer text configuration file.

The configuration file should be in Protocol Buffer text format (.textproto).
Templates can be loaded from a local directory or a Git repository.

Examples:
  custodian generate config.textproto
  custodian generate --template-dir ./templates config.textproto
  custodian generate --template-repo github.com/org/templates config.textproto
  custodian generate --output ./output --dry-run config.textproto`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.configFile = args[0]
			return runGenerate(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.outputDir, "output", "o", ".", "Output directory for generated Terraform files")
	cmd.Flags().StringVar(&opts.templateDir, "template-dir", "", "Local directory containing Terraform templates")
	cmd.Flags().StringVar(&opts.templateRepo, "template-repo", "", "Git repository URL containing Terraform templates")
	cmd.Flags().BoolVar(&opts.validate, "validate", true, "Validate configuration before generating")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show what would be generated without writing files")

	return cmd
}

func runGenerate(opts *generateOptions) error {
	// Read and parse the configuration file
	cfg, err := loadConfig(opts.configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration if requested
	if opts.validate {
		if err := validator.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
		fmt.Println("✓ Configuration validation passed")
	}

	// Determine template source
	templateSource := ""
	if opts.templateDir != "" {
		templateSource = opts.templateDir
	} else if opts.templateRepo != "" {
		templateSource = opts.templateRepo
	} else {
		// Use built-in templates
		templateSource = "builtin"
	}

	// Create generator
	gen, err := generator.New(templateSource)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Generate Terraform code
	files, err := gen.Generate(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate Terraform code: %w", err)
	}

	// Output results
	if opts.dryRun {
		fmt.Println("Files that would be generated:")
		for filename, content := range files {
			fmt.Printf("=== %s ===\n", filename)
			fmt.Println(content)
			fmt.Println()
		}
		return nil
	}

	// Write files to output directory
	for filename, content := range files {
		outputPath := filepath.Join(opts.outputDir, filename)
		if err := writeFile(outputPath, content); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}
		fmt.Printf("Generated: %s\n", outputPath)
	}

	fmt.Printf("✓ Generated %d Terraform files in %s\n", len(files), opts.outputDir)
	return nil
}

func loadConfig(filename string) (*config.Config, error) {
	content, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{}
	if err := prototext.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse Protocol Buffer text format: %w", err)
	}

	return cfg, nil
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}