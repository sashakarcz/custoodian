// Package generator provides functionality for generating Terraform code from protobuf configurations.
// It supports both built-in templates and custom templates from local directories or Git repositories.
package generator

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"custoodian/internal/templates"
	"custoodian/pkg/config"
)

// templateCacheEntry represents a cached template with metadata
type templateCacheEntry struct {
	templates *template.Template
	loadTime  time.Time
	source    string
}

// templateCache provides thread-safe caching of parsed templates
var (
	templateCache = make(map[string]*templateCacheEntry)
	cacheMutex    sync.RWMutex
	// cacheTimeout defines how long templates are cached (5 minutes for development)
	cacheTimeout = 5 * time.Minute
)

// Generator handles Terraform code generation from protobuf configurations.
// It manages template loading, parsing, and execution to produce infrastructure-as-code
// files that are compatible with Terraform for Google Cloud Platform resources.
//
// The generator supports multiple template sources:
//   - Built-in templates: Pre-defined templates for common GCP resources
//   - Local directory: Custom templates from a local filesystem directory
//   - Git repository: Custom templates from a remote Git repository (planned)
//
// Performance Optimizations:
//   - Template caching: Parsed templates are cached to avoid re-parsing
//   - Concurrent-safe: Multiple goroutines can safely use different Generator instances
//   - Memory efficient: Templates are shared when using the same source
//
// Thread Safety: Generator instances are not thread-safe for concurrent use,
// but multiple Generator instances can be used safely in different goroutines.
type Generator struct {
	// templateSource specifies where templates are loaded from.
	// Valid values: "builtin", local directory path, or Git repository URL
	templateSource string

	// templates holds the parsed template collection with custom functions
	// for converting protobuf enums to Terraform-compatible strings
	templates *template.Template

	// logger provides structured logging for debugging and monitoring
	logger *log.Logger
}

// NewOptions provides configuration options for creating a Generator
type NewOptions struct {
	// Logger provides custom logging. If nil, a default logger is used.
	Logger *log.Logger
	// DisableCache disables template caching for development/testing
	DisableCache bool
}

// New creates a new Generator instance with the specified template source.
//
// The templateSource parameter determines where templates are loaded from:
//   - "builtin" or empty string: Uses built-in templates embedded in the binary
//   - Local path: Loads templates from the specified directory (e.g., "./templates")
//   - Git URL: Loads templates from a Git repository (format: "github.com/org/repo")
//
// Returns an error if template loading fails, including cases where:
//   - Local directory doesn't exist or contains no .tf files
//   - Git repository is inaccessible (when implemented)
//   - Template parsing fails due to syntax errors
//
// Example usage:
//
//	gen, err := generator.New("builtin")
//	gen, err := generator.New("./custom-templates")
//	gen, err := generator.New("github.com/myorg/terraform-templates")
func New(templateSource string) (*Generator, error) {
	return NewWithOptions(templateSource, nil)
}

// NewWithOptions creates a new Generator with custom options.
//
// This function provides more control over Generator behavior, including
// custom logging and cache control. Use this when you need specific
// configuration for production or testing environments.
//
// Example usage:
//
//	opts := &NewOptions{
//	  Logger: customLogger,
//	  DisableCache: true, // for testing
//	}
//	gen, err := generator.NewWithOptions("builtin", opts)
func NewWithOptions(templateSource string, opts *NewOptions) (*Generator, error) {
	// Set up default options
	if opts == nil {
		opts = &NewOptions{}
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}

	g := &Generator{
		templateSource: templateSource,
		logger:         opts.Logger,
	}

	startTime := time.Now()
	if err := g.loadTemplates(!opts.DisableCache); err != nil {
		return nil, fmt.Errorf("failed to load templates from %s: %w", templateSource, err)
	}

	g.logger.Printf("Templates loaded from %s in %v", templateSource, time.Since(startTime))
	return g, nil
}

// Generate creates Terraform files from the given protobuf configuration.
//
// This method processes the entire configuration and generates a complete set of
// Terraform files organized by resource type. The generated files follow Terraform
// best practices and include proper resource dependencies.
//
// The generated file structure includes:
//   - project.tf: GCP project configuration, provider setup, and API enablement
//   - networking.tf: VPCs, subnets, firewall rules, NAT gateways, reserved IPs
//   - compute.tf: Instance templates, managed instance groups, individual instances
//   - load_balancers.tf: HTTP/HTTPS/TCP load balancers with health checks
//   - iam.tf: Service accounts, role bindings, custom roles
//   - storage.tf: Cloud Storage buckets with lifecycle policies
//   - variables.tf: Terraform input variables with sensible defaults
//   - outputs.tf: Terraform outputs for important resource attributes
//
// Parameters:
//   - cfg: The protobuf configuration containing all resource definitions
//
// Returns:
//   - map[string]string: A map where keys are filenames and values are file contents
//   - error: Any error encountered during generation
//
// The method will skip generating files for resource types that are not defined
// in the configuration (e.g., if no compute resources are specified, compute.tf
// won't be included in the result).
//
// Security Considerations:
//   - All string values are properly quoted to prevent injection attacks
//   - File paths are sanitized to prevent directory traversal
//   - Sensitive values (like service account keys) are marked as sensitive in outputs
func (g *Generator) Generate(cfg *config.Config) (map[string]string, error) {
	files := make(map[string]string)

	// Generate project configuration - this is required and includes provider setup
	if cfg.Project != nil {
		content, err := g.generateProject(cfg.Project)
		if err != nil {
			return nil, fmt.Errorf("failed to generate project configuration: %w", err)
		}
		files["project.tf"] = content
	}

	// Generate networking resources (VPCs, subnets, firewall rules, NAT gateways)
	if cfg.Networking != nil {
		content, err := g.generateNetworking(cfg.Networking)
		if err != nil {
			return nil, fmt.Errorf("failed to generate networking configuration: %w", err)
		}
		// Only include the file if it has actual content
		if content != "" {
			files["networking.tf"] = content
		}
	}

	// Generate compute resources (templates, instance groups, individual instances)
	if cfg.Compute != nil {
		content, err := g.generateCompute(cfg.Compute)
		if err != nil {
			return nil, fmt.Errorf("failed to generate compute configuration: %w", err)
		}
		if content != "" {
			files["compute.tf"] = content
		}
	}

	// Generate load balancer configurations with health checks
	if len(cfg.LoadBalancers) > 0 {
		content, err := g.generateLoadBalancers(cfg.LoadBalancers)
		if err != nil {
			return nil, fmt.Errorf("failed to generate load balancer configuration: %w", err)
		}
		files["load_balancers.tf"] = content
	}

	// Generate IAM resources (service accounts, role bindings, custom roles)
	if cfg.Iam != nil {
		content, err := g.generateIAM(cfg.Iam)
		if err != nil {
			return nil, fmt.Errorf("failed to generate IAM configuration: %w", err)
		}
		if content != "" {
			files["iam.tf"] = content
		}
	}

	// Generate storage resources (Cloud Storage buckets with lifecycle policies)
	if cfg.Storage != nil {
		content, err := g.generateStorage(cfg.Storage)
		if err != nil {
			return nil, fmt.Errorf("failed to generate storage configuration: %w", err)
		}
		if content != "" {
			files["storage.tf"] = content
		}
	}

	// Generate Cloud Run resources (services, VPC connectors)
	if cfg.CloudRun != nil {
		content, err := g.generateCloudRun(cfg.CloudRun)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Cloud Run configuration: %w", err)
		}
		if content != "" {
			files["cloud_run.tf"] = content
		}
	}

	// Generate database resources (Cloud SQL, Cloud Spanner)
	if cfg.Databases != nil {
		content, err := g.generateDatabases(cfg.Databases)
		if err != nil {
			return nil, fmt.Errorf("failed to generate database configuration: %w", err)
		}
		if content != "" {
			files["databases.tf"] = content
		}
	}

	// Generate variables file - always included with default values
	variables, err := g.generateVariables(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate variables configuration: %w", err)
	}
	files["variables.tf"] = variables

	// Generate outputs file - always included to expose important resource attributes
	outputs, err := g.generateOutputs(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate outputs configuration: %w", err)
	}
	files["outputs.tf"] = outputs

	return files, nil
}

// loadTemplates loads and parses templates from the specified source with optional caching.
//
// This method handles loading templates from three different sources:
//  1. Built-in templates: Embedded templates for standard GCP resources
//  2. Local directory: Templates from a filesystem directory
//  3. Git repository: Templates from a remote Git repository (not yet implemented)
//
// The method also sets up custom template functions that are available in all
// templates for converting protobuf enums to Terraform-compatible strings and
// performing common text transformations.
//
// Performance Features:
//   - Template caching: Avoids re-parsing templates when using the same source
//   - Cache invalidation: Templates expire after a configurable timeout
//   - Thread-safe: Multiple goroutines can safely access the cache
//
// Available template functions:
//   - regionToString: Converts Region enum to GCP region string (e.g., "us-central1")
//   - zoneToString: Converts Zone enum to GCP zone string (e.g., "us-central1-a")
//   - machineTypeToString: Converts MachineType enum to GCP machine type (e.g., "e2-medium")
//   - apiToString: Converts GcpApi enum to API service name (e.g., "compute.googleapis.com")
//   - networkTierToString: Converts NetworkTier enum to string (e.g., "PREMIUM")
//   - indent: Adds specified number of spaces to each line of text
//   - quote: Wraps string in double quotes for Terraform syntax
//   - join: Joins string slice with separator (strings.Join wrapper)
//   - lower/upper: String case conversion (strings.ToLower/ToUpper wrappers)
//   - replace: String replacement (strings.ReplaceAll wrapper)
//   - unescapeNewlines: Converts \n escape sequences to actual newlines
//
// Parameters:
//   - useCache: Whether to use template caching for performance
//
// Returns an error if:
//   - Template source cannot be accessed (directory doesn't exist, Git repo unreachable)
//   - Template parsing fails due to syntax errors
//   - No valid templates are found in the specified source
func (g *Generator) loadTemplates(useCache bool) error {
	// Check cache first if enabled
	if useCache {
		if cached := g.getCachedTemplate(); cached != nil {
			g.templates = cached
			g.logger.Printf("Using cached templates for source: %s", g.templateSource)
			return nil
		}
	}

	g.logger.Printf("Loading templates from source: %s", g.templateSource)

	var templateContent map[string]string
	var err error

	// Determine template source and load content
	switch g.templateSource {
	case "builtin", "":
		// Use embedded templates for standard GCP resources
		templateContent = templates.GetBuiltinTemplates()
		g.logger.Printf("Loaded %d built-in templates", len(templateContent))
	default:
		// Check if it's a local directory or a Git repository URL
		if strings.Contains(g.templateSource, "://") || strings.Contains(g.templateSource, "@") {
			// Git repository format detected (e.g., github.com/org/repo or git@github.com:org/repo.git)
			g.logger.Printf("Loading templates from Git repository: %s", g.templateSource)
			templateContent, err = templates.LoadFromGit(g.templateSource)
		} else {
			// Local directory path
			g.logger.Printf("Loading templates from directory: %s", g.templateSource)
			templateContent, err = templates.LoadFromDirectory(g.templateSource)
		}
		if err != nil {
			return fmt.Errorf("failed to load custom templates from %s: %w", g.templateSource, err)
		}
		g.logger.Printf("Loaded %d custom templates", len(templateContent))
	}

	// Initialize the template engine
	g.templates = template.New("custodian")

	// Register custom functions available to all templates
	g.templates = g.templates.Funcs(template.FuncMap{
		// GCP enum conversion functions
		"regionToString":      regionToString,
		"zoneToString":        zoneToString,
		"machineTypeToString": machineTypeToString,
		"apiToString":         apiToString,
		"networkTierToString": networkTierToString,

		// Text manipulation functions
		"indent":           indent,
		"quote":            quote,
		"join":             strings.Join,
		"lower":            strings.ToLower,
		"upper":            strings.ToUpper,
		"replace":          strings.ReplaceAll,
		"unescapeNewlines": func(s string) string { return strings.ReplaceAll(s, "\\n", "\n") },
	})

	// Parse each template and add it to the template collection
	templateCount := 0
	for name, content := range templateContent {
		if _, err := g.templates.New(name).Parse(content); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		templateCount++
	}

	g.logger.Printf("Successfully parsed %d templates", templateCount)

	// Cache the parsed templates if caching is enabled
	if useCache {
		g.cacheTemplate(g.templates)
	}

	return nil
}

// getCachedTemplate retrieves cached templates if they exist and are still valid
func (g *Generator) getCachedTemplate() *template.Template {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entry, exists := templateCache[g.templateSource]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(entry.loadTime) > cacheTimeout {
		// Cache expired, remove it
		go g.cleanExpiredCache()
		return nil
	}

	return entry.templates
}

// cacheTemplate stores parsed templates in the cache
func (g *Generator) cacheTemplate(templates *template.Template) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	templateCache[g.templateSource] = &templateCacheEntry{
		templates: templates,
		loadTime:  time.Now(),
		source:    g.templateSource,
	}

	g.logger.Printf("Cached templates for source: %s", g.templateSource)
}

// cleanExpiredCache removes expired entries from the template cache
func (g *Generator) cleanExpiredCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for source, entry := range templateCache {
		if now.Sub(entry.loadTime) > cacheTimeout {
			delete(templateCache, source)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		g.logger.Printf("Cleaned %d expired template cache entries", expiredCount)
	}
}

// generateProject generates Terraform configuration for GCP project setup.
//
// This includes the Terraform provider configuration, project resource creation,
// and API service enablement. The generated project.tf file serves as the
// foundation for all other resources.
//
// Generated resources:
//   - terraform and google provider configuration
//   - google_project resource with billing and organization setup
//   - google_project_service resources for each enabled API
func (g *Generator) generateProject(project *config.Project) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "project.tf", project)
	if err != nil {
		return "", fmt.Errorf("template execution failed for project configuration: %w", err)
	}
	return output.String(), nil
}

// TemplateContext provides comprehensive context for template execution with dependency information
type TemplateContext struct {
	// Primary data for the template
	Data interface{}
	// Dependency information
	Dependencies *DependencyInfo
}

// DependencyInfo contains information about resource dependencies
type DependencyInfo struct {
	// Whether project APIs need to be enabled first
	RequiresProjectAPIs bool
	// List of API dependencies
	ProjectAPIs []string
	// Whether networking resources are required
	RequiresNetworking bool
	// Network names that this resource depends on
	NetworkDependencies []string
}

// generateNetworking generates Terraform configuration for networking resources.
//
// This includes VPC networks, subnets, firewall rules, NAT gateways, and
// reserved IP addresses. Resources are organized hierarchically with proper
// dependencies (e.g., subnets reference their parent VPC).
//
// Generated resources:
//   - google_compute_address for reserved IPs
//   - google_compute_network for VPC networks
//   - google_compute_subnetwork for subnets with secondary ranges
//   - google_compute_firewall for firewall rules
//   - google_compute_router_nat for NAT gateways
func (g *Generator) generateNetworking(networking *config.Networking) (string, error) {
	// Create template context with dependency information
	ctx := &TemplateContext{
		Data: networking,
		Dependencies: &DependencyInfo{
			RequiresProjectAPIs: true,
			ProjectAPIs:         []string{"compute.googleapis.com"},
			RequiresNetworking:  false, // This IS the networking layer
		},
	}
	
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "networking.tf", ctx)
	if err != nil {
		return "", fmt.Errorf("template execution failed for networking configuration: %w", err)
	}
	return output.String(), nil
}

// generateCompute generates Terraform configuration for compute resources.
//
// This includes instance templates, managed instance groups, autoscalers,
// and individual VM instances. Instance groups automatically reference
// their templates and include autoscaling policies when specified.
//
// Generated resources:
//   - google_compute_instance_template with disks, networking, and metadata
//   - google_compute_instance_group_manager for managed groups
//   - google_compute_autoscaler for auto-scaling policies
//   - google_compute_instance for individual VMs
func (g *Generator) generateCompute(compute *config.Compute) (string, error) {
	// Collect network dependencies from compute configuration
	var networkDeps []string
	
	// Check instance templates for network dependencies
	for _, template := range compute.InstanceTemplates {
		for _, netIface := range template.NetworkInterfaces {
			if netIface.Network != "" {
				networkDeps = append(networkDeps, fmt.Sprintf("google_compute_network.%s", netIface.Network))
			}
			if netIface.Subnetwork != "" {
				networkDeps = append(networkDeps, fmt.Sprintf("google_compute_subnetwork.%s", netIface.Subnetwork))
			}
		}
	}
	
	// Check individual instances for network dependencies
	for _, instance := range compute.Instances {
		for _, netIface := range instance.NetworkInterfaces {
			if netIface.Network != "" {
				networkDeps = append(networkDeps, fmt.Sprintf("google_compute_network.%s", netIface.Network))
			}
			if netIface.Subnetwork != "" {
				networkDeps = append(networkDeps, fmt.Sprintf("google_compute_subnetwork.%s", netIface.Subnetwork))
			}
		}
	}
	
	// Create template context with dependency information
	ctx := &TemplateContext{
		Data: compute,
		Dependencies: &DependencyInfo{
			RequiresProjectAPIs:     true,
			ProjectAPIs:            []string{"compute.googleapis.com"},
			RequiresNetworking:     len(networkDeps) > 0,
			NetworkDependencies:    networkDeps,
		},
	}
	
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "compute.tf", ctx)
	if err != nil {
		return "", fmt.Errorf("template execution failed for compute configuration: %w", err)
	}
	return output.String(), nil
}

// generateLoadBalancers generates Terraform configuration for load balancers.
//
// This creates complete load balancing setups including forwarding rules,
// target proxies, URL maps, backend services, and health checks. The
// configuration supports HTTP, HTTPS, and TCP load balancers.
//
// Generated resources:
//   - google_compute_global_forwarding_rule for traffic entry points
//   - google_compute_target_http_proxy for HTTP load balancers
//   - google_compute_url_map for routing rules
//   - google_compute_backend_service for backend configuration
//   - google_compute_health_check for health monitoring
func (g *Generator) generateLoadBalancers(lbs []*config.LoadBalancer) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "load_balancers.tf", lbs)
	if err != nil {
		return "", fmt.Errorf("template execution failed for load balancer configuration: %w", err)
	}
	return output.String(), nil
}

// generateIAM generates Terraform configuration for IAM resources.
//
// This includes service accounts, role bindings, and custom roles.
// Service accounts can be configured with automatic key generation
// and role assignments. Role bindings support conditional access.
//
// Generated resources:
//   - google_service_account for service identities
//   - google_service_account_key for authentication keys (when requested)
//   - google_project_iam_member for individual role assignments
//   - google_project_iam_binding for group role assignments
//   - google_project_iam_custom_role for custom role definitions
func (g *Generator) generateIAM(iam *config.Iam) (string, error) {
	var output strings.Builder
	
	// Create template context with dependencies
	ctx := &TemplateContext{
		Data: iam,
		Dependencies: &DependencyInfo{
			RequiresProjectAPIs:     false,
			ProjectAPIs:            []string{},
			RequiresNetworking:     false,
			NetworkDependencies:    []string{},
		},
	}
	
	err := g.templates.ExecuteTemplate(&output, "iam.tf", ctx)
	if err != nil {
		return "", fmt.Errorf("template execution failed for IAM configuration: %w", err)
	}
	return output.String(), nil
}

// generateStorage generates Terraform configuration for storage resources.
//
// This includes Cloud Storage buckets with comprehensive configuration
// including lifecycle policies, versioning, and access controls.
//
// Generated resources:
//   - google_storage_bucket with location, storage class, and access settings
//   - Lifecycle rules for automatic storage class transitions and deletion
//   - Versioning and uniform bucket-level access configuration
func (g *Generator) generateStorage(storage *config.Storage) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "storage.tf", storage)
	if err != nil {
		return "", fmt.Errorf("template execution failed for storage configuration: %w", err)
	}
	return output.String(), nil
}

// generateVariables generates the variables.tf file with input variable definitions.
//
// This file defines Terraform input variables with sensible defaults,
// making the generated infrastructure reusable across different environments.
// Variables include project ID, region, and zone settings.
//
// Generated variables:
//   - project_id: GCP project identifier
//   - region: Default GCP region for regional resources
//   - zone: Default GCP zone for zonal resources
func (g *Generator) generateVariables(cfg *config.Config) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "variables.tf", cfg)
	if err != nil {
		return "", fmt.Errorf("template execution failed for variables configuration: %w", err)
	}
	return output.String(), nil
}

// generateOutputs generates the outputs.tf file with resource output values.
//
// This file exposes important attributes of created resources, making them
// available for reference by other Terraform configurations or for display
// to users. Sensitive outputs (like service account keys) are marked appropriately.
//
// Generated outputs:
//   - Project ID and number
//   - Network and subnet IDs and self-links
//   - Reserved IP addresses
//   - Service account emails and keys (sensitive)
func (g *Generator) generateOutputs(cfg *config.Config) (string, error) {
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "outputs.tf", cfg)
	if err != nil {
		return "", fmt.Errorf("template execution failed for outputs configuration: %w", err)
	}
	return output.String(), nil
}

// generateCloudRun generates Terraform configuration for Cloud Run resources.
//
// This includes Cloud Run services with comprehensive configuration including
// container settings, environment variables, secrets, traffic allocation,
// and IAM bindings. Also supports VPC Access Connectors for private networking.
//
// Generated resources:
//   - google_cloud_run_service for containerized applications
//   - google_cloud_run_service_iam_member for access control
//   - google_vpc_access_connector for VPC connectivity
func (g *Generator) generateCloudRun(cloudRun *config.CloudRun) (string, error) {
	// Create template context with dependency information
	ctx := &TemplateContext{
		Data: cloudRun,
		Dependencies: &DependencyInfo{
			RequiresProjectAPIs:     true,
			ProjectAPIs:            []string{"run.googleapis.com", "vpcaccess.googleapis.com"},
			RequiresNetworking:     false, // Cloud Run doesn't directly depend on networking resources
			NetworkDependencies:    []string{},
		},
	}
	
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "cloud_run.tf", ctx)
	if err != nil {
		return "", fmt.Errorf("template execution failed for Cloud Run configuration: %w", err)
	}
	return output.String(), nil
}

// generateDatabases generates Terraform configuration for database resources.
//
// This includes Cloud SQL instances with comprehensive configuration including
// storage, networking, backup, high availability, and database/user management.
// Also supports Cloud Spanner instances with databases and schema definitions.
//
// Generated resources:
//   - google_sql_database_instance for managed relational databases
//   - google_sql_database for individual databases within instances
//   - google_sql_user for database users and authentication
//   - google_spanner_instance for globally distributed databases
//   - google_spanner_database for Spanner databases with DDL schema
func (g *Generator) generateDatabases(databases *config.Databases) (string, error) {
	// Create template context with dependency information
	ctx := &TemplateContext{
		Data: databases,
		Dependencies: &DependencyInfo{
			RequiresProjectAPIs:     true,
			ProjectAPIs:            []string{"sqladmin.googleapis.com", "spanner.googleapis.com"},
			RequiresNetworking:     false, // Database networking is separate from VPC resources
			NetworkDependencies:    []string{},
		},
	}
	
	var output strings.Builder
	err := g.templates.ExecuteTemplate(&output, "databases.tf", ctx)
	if err != nil {
		return "", fmt.Errorf("template execution failed for database configuration: %w", err)
	}
	return output.String(), nil
}
