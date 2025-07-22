package generator

import (
	"fmt"
	"strings"
	"text/template"

	"custoodian/internal/templates"
	"custoodian/pkg/config"
)

// Generator handles Terraform code generation from protobuf configurations
type Generator struct {
	templateSource string
	templates      *template.Template
}

// New creates a new Generator with the specified template source
func New(templateSource string) (*Generator, error) {
	g := &Generator{
		templateSource: templateSource,
	}

	if err := g.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return g, nil
}

// Generate creates Terraform files from the given configuration
func (g *Generator) Generate(cfg *config.Config) (map[string]string, error) {
	files := make(map[string]string)
	
	// Generate project configuration
	if cfg.Project != nil {
		content, err := g.generateProject(cfg.Project)
		if err != nil {
			return nil, fmt.Errorf("failed to generate project: %w", err)
		}
		files["project.tf"] = content
	}

	// Generate networking resources
	if cfg.Networking != nil {
		content, err := g.generateNetworking(cfg.Networking)
		if err != nil {
			return nil, fmt.Errorf("failed to generate networking: %w", err)
		}
		if content != "" {
			files["networking.tf"] = content
		}
	}

	// Generate compute resources
	if cfg.Compute != nil {
		content, err := g.generateCompute(cfg.Compute)
		if err != nil {
			return nil, fmt.Errorf("failed to generate compute: %w", err)
		}
		if content != "" {
			files["compute.tf"] = content
		}
	}

	// Generate load balancers
	if len(cfg.LoadBalancers) > 0 {
		content, err := g.generateLoadBalancers(cfg.LoadBalancers)
		if err != nil {
			return nil, fmt.Errorf("failed to generate load balancers: %w", err)
		}
		files["load_balancers.tf"] = content
	}

	// Generate IAM resources
	if cfg.Iam != nil {
		content, err := g.generateIAM(cfg.Iam)
		if err != nil {
			return nil, fmt.Errorf("failed to generate IAM: %w", err)
		}
		if content != "" {
			files["iam.tf"] = content
		}
	}

	// Generate storage resources
	if cfg.Storage != nil {
		content, err := g.generateStorage(cfg.Storage)
		if err != nil {
			return nil, fmt.Errorf("failed to generate storage: %w", err)
		}
		if content != "" {
			files["storage.tf"] = content
		}
	}

	// Generate variables file
	variables, err := g.generateVariables(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate variables: %w", err)
	}
	files["variables.tf"] = variables

	// Generate outputs file
	outputs, err := g.generateOutputs(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate outputs: %w", err)
	}
	files["outputs.tf"] = outputs

	return files, nil
}

// loadTemplates loads templates from the specified source
func (g *Generator) loadTemplates() error {
	var templateContent map[string]string
	var err error

	switch g.templateSource {
	case "builtin", "":
		templateContent = templates.GetBuiltinTemplates()
	default:
		// Check if it's a local directory or a Git repository
		if strings.Contains(g.templateSource, "://") || strings.Contains(g.templateSource, "@") {
			templateContent, err = templates.LoadFromGit(g.templateSource)
		} else {
			templateContent, err = templates.LoadFromDirectory(g.templateSource)
		}
		if err != nil {
			return err
		}
	}

	// Parse templates
	g.templates = template.New("custodian")
	
	// Add custom functions
	g.templates = g.templates.Funcs(template.FuncMap{
		"regionToString":     regionToString,
		"zoneToString":       zoneToString,
		"machineTypeToString": machineTypeToString,
		"apiToString":        apiToString,
		"indent":             indent,
		"quote":              quote,
		"join":               strings.Join,
		"lower":              strings.ToLower,
		"upper":              strings.ToUpper,
		"replace":            strings.ReplaceAll,
		"unescapeNewlines":   func(s string) string { return strings.ReplaceAll(s, "\\n", "\n") },
	})

	// Parse each template
	for name, content := range templateContent {
		if _, err := g.templates.New(name).Parse(content); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
	}

	return nil
}

// generateProject generates Terraform code for project configuration
func (g *Generator) generateProject(project *config.Project) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "project.tf", project)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateNetworking generates Terraform code for networking resources
func (g *Generator) generateNetworking(networking *config.Networking) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "networking.tf", networking)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateCompute generates Terraform code for compute resources
func (g *Generator) generateCompute(compute *config.Compute) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "compute.tf", compute)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateLoadBalancers generates Terraform code for load balancers
func (g *Generator) generateLoadBalancers(lbs []*config.LoadBalancer) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "load_balancers.tf", lbs)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateIAM generates Terraform code for IAM resources
func (g *Generator) generateIAM(iam *config.Iam) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "iam.tf", iam)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateStorage generates Terraform code for storage resources
func (g *Generator) generateStorage(storage *config.Storage) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "storage.tf", storage)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateVariables generates the variables.tf file
func (g *Generator) generateVariables(cfg *config.Config) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "variables.tf", cfg)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateOutputs generates the outputs.tf file
func (g *Generator) generateOutputs(cfg *config.Config) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "outputs.tf", cfg)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}