package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type schemaOptions struct {
	format string
	output string
}

func newSchemaCmd() *cobra.Command {
	opts := &schemaOptions{
		format: "proto",
	}

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Display or export the Protocol Buffer schema",
		Long: `Display or export the Protocol Buffer schema used for configuration.

This command can output the schema in different formats for documentation
or IDE integration purposes.

Examples:
  custodian schema                    # Display proto schema
  custodian schema --format json     # Display JSON schema
  custodian schema --output schema/  # Export to directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSchema(opts)
		},
	}

	cmd.Flags().StringVar(&opts.format, "format", "proto", "Output format (proto, json, markdown)")
	cmd.Flags().StringVar(&opts.output, "output", "", "Output directory (default: stdout)")

	return cmd
}

func runSchema(opts *schemaOptions) error {
	switch opts.format {
	case "proto":
		return outputProtoSchema(opts.output)
	case "json":
		return outputJSONSchema(opts.output)
	case "markdown":
		return outputMarkdownSchema(opts.output)
	default:
		return fmt.Errorf("unsupported format: %s", opts.format)
	}
}

func outputProtoSchema(output string) error {
	schemas := map[string]string{
		"config.proto": getConfigProtoContent(),
		"enums.proto":  getEnumsProtoContent(),
	}

	if output == "" {
		for filename, content := range schemas {
			fmt.Printf("=== %s ===\n", filename)
			fmt.Println(content)
			fmt.Println()
		}
		return nil
	}

	for filename, content := range schemas {
		if err := writeFile(fmt.Sprintf("%s/%s", output, filename), content); err != nil {
			return err
		}
		fmt.Printf("Exported: %s/%s\n", output, filename)
	}

	return nil
}

func outputJSONSchema(output string) error {
	// TODO: Implement JSON schema generation
	return fmt.Errorf("JSON schema format not yet implemented")
}

func outputMarkdownSchema(output string) error {
	// TODO: Implement Markdown documentation generation
	return fmt.Errorf("Markdown schema format not yet implemented")
}

func getConfigProtoContent() string {
	// This would typically read from embedded files or generate from protobuf descriptors
	return `// Protocol Buffer schema for GCP infrastructure configuration
// See proto/custodian/config.proto for the full definition`
}

func getEnumsProtoContent() string {
	return `// Protocol Buffer enums for GCP resources
// See proto/custodian/enums.proto for the full definition`
}

func init() {
	rootCmd.AddCommand(newSchemaCmd())
}