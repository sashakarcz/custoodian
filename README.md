# Custoodian - GCP Terraform Generator

[![Build Status](https://github.com/custoodian/custoodian/workflows/CI/badge.svg)](https://github.com/custoodian/custoodian/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/custoodian/custoodian)](https://goreportcard.com/report/github.com/custoodian/custoodian)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Custoodian is a tool that generates Terraform code from Protocol Buffer text configurations for Google Cloud Platform resources. It provides type-safe infrastructure configuration with comprehensive validation and supports custom template systems.

Custoodian leverages Protocol Buffers for strong typing and validation, catching configuration errors before Terraform runs.

## ✨ Features

- **Protocol Buffer-based Configuration**: Type-safe infrastructure definitions with compile-time validation
- **Comprehensive GCP Support**: Full coverage of GCP resources including compute, networking, storage, Cloud Run, and IAM
- **Template System**: Built-in templates with support for custom template directories and Git repositories
- **Rich Validation**: Extensive validation rules using proto-validate extensions
- **CLI Tool**: Fast, standalone binary that works locally or in CI/CD pipelines
- **GitHub Action**: Ready-to-use GitHub Action for CI/CD integration
- **VS Code Integration**: Seamless development experience with Protocol Buffer language server

## 🚀 Quick Start

### Installation

#### Download Binary

```bash
# Linux/macOS
curl -L -o /usr/local/bin/custoodian https://github.com/custoodian/custoodian/releases/latest/download/custoodian-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')
chmod +x /usr/local/bin/custoodian
```

#### Build from Source

```bash
git clone https://github.com/custoodian/custoodian.git
cd custoodian
make build
sudo make install
```

### Basic Usage

1. **Create a configuration file** (`config.textproto`):

```protobuf
project {
  id: "my-project-123"
  name: "My Application"
  billing_account: "123456-ABCDEF-GHIJKL"
  apis: [GCP_API_COMPUTE, GCP_API_STORAGE, GCP_API_IAM]
}

networking {
  reserved_ips {
    name: "web-lb-ip"
    type: RESERVED_IP_TYPE_GLOBAL
  }
  
  vpcs {
    name: "main-vpc"
    subnets {
      name: "web-subnet"
      cidr: "10.0.1.0/24"
      region: REGION_US_CENTRAL1
    }
  }
}

compute {
  instance_templates {
    name: "web-template"
    machine_type: MACHINE_TYPE_E2_MEDIUM
    image: "ubuntu-2004-lts"
    disk_size_gb: 20
    
    network_interfaces {
      subnetwork: "web-subnet"
      access_configs {
        type: "ONE_TO_ONE_NAT"
      }
    }
  }
  
  instance_groups {
    name: "web-group"
    template: "web-template"
    size: 3
    zones: [ZONE_US_CENTRAL1_A, ZONE_US_CENTRAL1_B]
    
    auto_scaling {
      min: 2
      max: 10
      cpu_target: 0.6
    }
  }
}

load_balancers {
  name: "main-lb"
  type: LOAD_BALANCER_TYPE_HTTP
  ip: "web-lb-ip"
  backend: "web-group"
}
```

2. **Validate the configuration**:

```bash
custoodian validate config.textproto
```

3. **Generate Terraform code**:

```bash
custoodian generate config.textproto --output ./terraform
```

4. **Apply with Terraform**:

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### Cloud Run Example

For serverless containerized applications, Custoodian supports comprehensive Cloud Run configuration:

```protobuf
project {
  id: "my-serverless-app"
  billing_account: "123456-ABCDEF-GHIJKL"
  apis: [GCP_API_CLOUD_RUN, GCP_API_VPC_ACCESS, GCP_API_IAM]
}

cloud_run {
  services {
    name: "hello-world"
    description: "Simple web service"
    location: REGION_US_CENTRAL1
    image: "gcr.io/my-project/hello-world:latest"
    
    config {
      port: 8080
      cpu: "1000m"
      memory: "512Mi"
      max_instances: 100
      min_instances: 0
      max_concurrent_requests: 80
      timeout_seconds: 300
      
      env_vars: {
        key: "NODE_ENV"
        value: "production"
      }
      
      env_from_secrets {
        name: "DATABASE_URL"
        secret_name: "db-connection"
        version: "latest"
      }
    }
    
    traffic {
      percent: 100
    }
    
    iam_bindings {
      role: "roles/run.invoker"
      members: ["allUsers"]
    }
  }
  
  vpc_connectors {
    name: "serverless-vpc-connector"
    network: "default"
    ip_cidr_range: "10.8.0.0/28"
    min_instances: 2
    max_instances: 10
  }
}
```

This generates Cloud Run services with:
- **Resource Limits**: CPU, memory, and scaling configuration
- **Environment Variables**: From values and Google Secret Manager
- **Traffic Management**: Blue/green deployments with percentage splits
- **IAM Access Control**: Service-level permissions
- **VPC Connectivity**: Private networking with VPC Access Connectors

## 📖 Documentation

### Protocol Buffer Schema

Custoodian uses Protocol Buffers to define infrastructure configurations. The main message types include:

- `Project`: GCP project configuration, APIs, billing
- `Networking`: VPCs, subnets, firewall rules, NAT gateways, reserved IPs
- `Compute`: Instance templates, managed instance groups, individual instances
- `LoadBalancer`: HTTP/HTTPS/TCP load balancers with health checks
- `Iam`: Service accounts, role bindings, custom roles
- `Storage`: Cloud Storage buckets with lifecycle policies
- `CloudRun`: Containerized services, VPC connectors, IAM bindings

### Field Validation

All fields include comprehensive validation rules:

```protobuf
message Project {
  // Project ID must be 6-30 characters, lowercase, start with letter
  string id = 1 [(buf.validate.field).string.pattern = "^[a-z][a-z0-9-]{4,28}[a-z0-9]$"];
  
  // Billing account format: 123456-ABCDEF-GHIJKL  
  string billing_account = 3 [(buf.validate.field).string.pattern = "^[0-9]{6}-[A-Z0-9]{6}-[A-Z0-9]{6}$"];
}
```

### Resource Enums

All GCP-specific values use strongly-typed enums:

```protobuf
enum Region {
  REGION_US_CENTRAL1 = 1;
  REGION_US_EAST1 = 2;
  REGION_EUROPE_WEST1 = 8;
  // ... more regions
}

enum MachineType {
  MACHINE_TYPE_E2_MICRO = 1;
  MACHINE_TYPE_E2_MEDIUM = 3;
  MACHINE_TYPE_N1_STANDARD_4 = 10;
  // ... more machine types
}
```

### CLI Commands

#### Generate Terraform Code

```bash
# Basic generation
custoodian generate config.textproto

# With custom output directory
custoodian generate config.textproto --output ./infrastructure

# Using custom templates
custoodian generate config.textproto --template-dir ./custom-templates

# Using templates from Git repository
custoodian generate config.textproto --template-repo github.com/org/templates

# Dry run (show what would be generated)
custoodian generate config.textproto --dry-run
```

#### Validate Configuration

```bash
# Validate syntax and constraints
custoodian validate config.textproto
```

#### Display Schema

```bash
# Show Protocol Buffer schema
custoodian schema

# Export schema to directory
custoodian schema --output ./schema
```

### Custom Templates

Custoodian supports custom Terraform templates for organizations that need specific patterns:

#### Local Template Directory

```bash
custoodian generate config.textproto --template-dir ./my-templates
```

Template directory structure:
```
my-templates/
├── project.tf
├── networking.tf  
├── compute.tf
├── load_balancers.tf
├── iam.tf
├── storage.tf
├── cloud_run.tf
├── variables.tf
└── outputs.tf
```

#### Git Repository Templates

```bash
custoodian generate config.textproto --template-repo github.com/myorg/gcp-templates
```

## 📝 Creating Custom Templates

Custom templates allow you to customize the generated Terraform code to match your organization's standards, naming conventions, and specific requirements.

### Template Structure

Each template file corresponds to a specific resource type and receives structured data from the Protocol Buffer configuration:

| Template File | Data Type | Purpose |
|---------------|-----------|---------|
| `project.tf` | `*config.Project` | GCP project, provider, APIs |
| `networking.tf` | `TemplateContext{Data: *config.Networking}` | VPCs, subnets, firewall rules |
| `compute.tf` | `TemplateContext{Data: *config.Compute}` | VMs, instance groups, templates |
| `load_balancers.tf` | `[]*config.LoadBalancer` | Load balancers, health checks |
| `iam.tf` | `*config.Iam` | Service accounts, role bindings |
| `storage.tf` | `*config.Storage` | Cloud Storage buckets |
| `cloud_run.tf` | `TemplateContext{Data: *config.CloudRun}` | Containerized services, VPC connectors |
| `variables.tf` | `*config.Config` | Terraform input variables |
| `outputs.tf` | `*config.Config` | Terraform output values |

### Template Context System

Templates that support dependency management receive a `TemplateContext` object:

```go
type TemplateContext struct {
    Data         interface{}      // The actual resource data
    Dependencies *DependencyInfo  // Dependency metadata
}

type DependencyInfo struct {
    RequiresProjectAPIs bool     // Whether APIs need to be enabled first
    ProjectAPIs         []string // List of required API services
    RequiresNetworking  bool     // Whether networking resources are needed
    NetworkDependencies []string // Network resource references
}
```

### Available Template Functions

Custoodian provides helper functions for common operations:

```go
// String manipulation
quote(s string) string              // Safely quote strings
indent(spaces int, text string)     // Indent text blocks
unescapeNewlines(s string) string   // Process startup scripts

// GCP-specific conversions
regionToString(region Region) string           // Convert region enum
zoneToString(zone Zone) string                // Convert zone enum  
machineTypeToString(mt MachineType) string    // Convert machine type
networkTierToString(nt NetworkTier) string    // Convert network tier
```

### Example: Custom Networking Template

Here's how to create a custom networking template with organization-specific patterns:

```hcl
# networking.tf
{{- $data := .Data -}}
{{- $deps := .Dependencies -}}

# Organization: ACME Corp
# Template Version: 2.1
# Generated: {{ "{{ .TimeStamp }}" }}

{{- if $data.Vpcs}}
# VPC Networks
{{- range $data.Vpcs}}
resource "google_compute_network" "{{ .Name }}" {
  name                    = "acme-{{ .Name }}-{{ "{{ var.environment }}" }}"
  description             = "{{ .Description }} - Managed by Custoodian"
  auto_create_subnetworks = {{ .AutoCreateSubnetworks }}
  routing_mode            = {{ quote (upper .RoutingMode) }}
  
  {{- if $deps.RequiresProjectAPIs}}
  depends_on = [
    {{- range $i, $api := $deps.ProjectAPIs}}
    {{- if $i}},{{end}}
    google_project_service.{{ $api | replace "." "_" }}
    {{- end}}
  ]
  {{- end}}

  # ACME Corp standard labels
  labels = {
    environment    = var.environment
    cost_center    = var.cost_center
    managed_by     = "custoodian"
    creation_date  = "{{ "{{ formatdate("YYYY-MM-DD", timestamp()) }}" }}"
    {{- if .Labels}}
    {{- range $key, $value := .Labels}}
    {{ quote $key }} = {{ quote $value }}
    {{- end}}
    {{- end}}
  }
}

{{- if .Subnets}}
# Subnets for {{ .Name }}
{{- range .Subnets}}
resource "google_compute_subnetwork" "{{ .Name }}" {
  name          = "acme-{{ .Name }}-{{ "{{ var.environment }}" }}"
  description   = "{{ .Description }} - Auto-managed"
  ip_cidr_range = {{ quote .Cidr }}
  region        = {{ quote (regionToString .Region) }}
  network       = google_compute_network.{{ $.Name }}.id
  
  # ACME Corp requires private Google access
  private_ip_google_access = true
  
  log_config {
    aggregation_interval = "INTERVAL_10_MIN"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}
{{- end}}
{{- end}}
{{- end}}
{{- end}}

# Organization-wide firewall rules
resource "google_compute_firewall" "acme_deny_all" {
  name        = "acme-deny-all-{{ "{{ var.environment }}" }}"
  description = "ACME Corp security baseline - deny all by default"
  network     = google_compute_network.{{ (index $data.Vpcs 0).Name }}.name
  direction   = "INGRESS"
  priority    = 65534

  deny {
    protocol = "all"
  }

  source_ranges = ["0.0.0.0/0"]
}
```

### Example: Custom Variables Template

Create standardized variables for your organization:

```hcl
# variables.tf
# ACME Corp Standard Variables
# Generated by Custoodian

variable "project_id" {
  description = "The GCP project ID"
  type        = string
  
  validation {
    condition     = can(regex("^acme-", var.project_id))
    error_message = "Project ID must start with 'acme-' prefix."
  }
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "cost_center" {
  description = "Cost center for billing allocation"
  type        = string
  
  validation {
    condition     = can(regex("^CC-[0-9]{4}$", var.cost_center))
    error_message = "Cost center must match format CC-NNNN."
  }
}

variable "region" {
  description = "Default GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "Default GCP zone" 
  type        = string
  default     = "us-central1-a"
}

# Dynamic variables based on configuration
{{- if .Networking}}
{{- if .Networking.Vpcs}}
variable "vpc_cidrs" {
  description = "CIDR blocks for VPC networks"
  type        = map(string)
  default = {
    {{- range .Networking.Vpcs}}
    {{- if .Subnets}}
    {{- range .Subnets}}
    "{{ .Name }}" = {{ quote .Cidr }}
    {{- end}}
    {{- end}}
    {{- end}}
  }
}
{{- end}}
{{- end}}
```

### Template Development Workflow

1. **Start with built-in templates**: Copy the built-in templates as a starting point
```bash
# View built-in template (use this as reference)
custoodian generate --dry-run config.textproto
```

2. **Create template directory**: Set up your custom template directory
```bash
mkdir ./my-templates
cd my-templates
```

3. **Develop iteratively**: Test changes frequently
```bash
# Test template changes
custoodian generate config.textproto --template-dir ./my-templates --dry-run

# Generate to files when satisfied
custoodian generate config.textproto --template-dir ./my-templates -o ./output
```

4. **Validate output**: Always validate generated Terraform
```bash
cd ./output
terraform init
terraform validate
terraform plan
```

### Advanced Template Patterns

#### Conditional Resource Creation
```hcl
{{- if .EnableMonitoring}}
resource "google_monitoring_alert_policy" "high_cpu" {
  display_name = "High CPU Usage"
  # ... monitoring configuration
}
{{- end}}
```

#### Loop with Complex Logic
```hcl
{{- range .Instances}}
{{- $instance := . }}
resource "google_compute_instance" "{{ .Name }}" {
  name = {{ quote .Name }}
  
  {{- if .NetworkInterfaces}}
  {{- range .NetworkInterfaces}}
  network_interface {
    {{- if .Network}}
    # Use resource reference for networks defined in this config
    {{- $networkFound := false }}
    {{- range $data.Vpcs }}
      {{- if eq .Name $.Network }}{{- $networkFound = true }}{{- end }}
    {{- end }}
    {{- if $networkFound }}
    network = google_compute_network.{{ .Network }}.id
    {{- else }}
    network = {{ quote .Network }}
    {{- end }}
    {{- end}}
  }
  {{- end}}
  {{- end}}
  
  {{- if $deps.RequiresNetworking}}
  depends_on = [
    {{- range $deps.NetworkDependencies}}
    {{ . }},
    {{- end}}
  ]
  {{- end}}
}
{{- end}}
```

#### Organization Policies
```hcl
# Add organization-specific resources
resource "google_project_organization_policy" "restrict_vm_external_ips" {
  project    = var.project_id
  constraint = "compute.vmExternalIpAccess"

  list_policy {
    deny {
      all = true
    }
  }
}
```

### Best Practices

1. **Use Template Comments**: Document your template logic
2. **Validate Input**: Add validation blocks to variables
3. **Follow Naming Conventions**: Use consistent resource naming
4. **Handle Edge Cases**: Check for empty values and optional fields
5. **Test Thoroughly**: Validate generated Terraform in multiple scenarios
6. **Version Control**: Keep templates in Git with proper versioning
7. **Use Dependencies**: Leverage the dependency system for proper resource ordering

### Troubleshooting Templates

```bash
# Debug template execution
CUSTOODIAN_LOG_LEVEL=debug custoodian generate config.textproto --template-dir ./templates

# Test specific template
custoodian generate config.textproto --template-dir ./templates --dry-run | grep -A 20 "networking.tf"

# Validate template syntax
custoodian generate test-config.textproto --template-dir ./templates -o /tmp/test-output
cd /tmp/test-output && terraform validate
```

### GitHub Action

Use Custoodian in your CI/CD pipeline:

```yaml
name: Generate Infrastructure
on:
  pull_request:
    paths:
      - 'infrastructure/config.textproto'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Generate Terraform
        uses: custoodian/custoodian@v1
        with:
          config-file: 'infrastructure/config.textproto'
          output-dir: './generated-terraform'
          
      - name: Upload Terraform files
        uses: actions/upload-artifact@v3
        with:
          name: terraform-files
          path: ./generated-terraform
```

Action inputs:
- `config-file`: Path to Protocol Buffer configuration file (required)
- `output-dir`: Output directory for Terraform files (default: `./terraform`)
- `template-dir`: Local template directory (optional)
- `template-repo`: Git repository with templates (optional)
- `validate-only`: Only validate, don't generate (default: `false`)
- `dry-run`: Show what would be generated (default: `false`)

## 🔧 VS Code Integration

For the best development experience, install the Protocol Buffers extension:

1. Install [vscode-proto3](https://marketplace.visualstudio.com/items?itemName=zxh404.vscode-proto3)
2. Add to your workspace settings:

```json
{
  "protoc": {
    "options": [
      "--proto_path=${workspaceFolder}/proto"
    ]
  }
}
```

3. Create `.vscode/settings.json`:

```json
{
  "files.associations": {
    "*.textproto": "proto3"
  }
}
```

This provides:
- Syntax highlighting for `.textproto` files
- Auto-completion based on proto schema
- Real-time validation and error checking
- Go-to-definition for message types

## 🏗️ Examples

### Simple Web Application

See [`examples/simple.textproto`](examples/simple.textproto) for a basic web application with:
- Project with enabled APIs
- VPC with subnets and firewall rules  
- Auto-scaling compute instances
- HTTP load balancer
- IAM service accounts
- Cloud Storage bucket

### Advanced Enterprise Setup

See [`examples/advanced.textproto`](examples/advanced.textproto) for an enterprise-grade configuration with:
- Multi-region deployment
- Private instances with NAT gateways
- HTTPS load balancer with SSL
- Comprehensive IAM setup
- Multiple storage tiers with lifecycle policies
- Advanced networking and security

## 🏗️ Architecture

### Core Components

Custoodian follows a modular architecture designed for extensibility and maintainability:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Validation      │    │   Generation    │
│                 │    │                  │    │                 │
│ • Command       │────▶ • Proto         │────▶ • Template      │
│   Parsing       │    │   Validation     │    │   Processing    │
│ • Flag          │    │ • Business       │    │ • Resource      │
│   Handling      │    │   Rules          │    │   Generation    │
│ • File I/O      │    │ • Cross-refs     │    │ • Optimization  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Configuration   │    │    Templates     │    │     Output      │
│                 │    │                  │    │                 │
│ • Protobuf      │    │ • Built-in       │    │ • Terraform     │
│   Parsing       │    │ • Local Dir      │    │   Files         │
│ • Validation    │    │ • Git Repos      │    │ • Validation    │
│ • Type Safety   │    │ • Caching        │    │ • Dependencies  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Performance Features

- **Template Caching**: Parsed templates are cached in memory with configurable TTL
- **Concurrent Safety**: Thread-safe template cache with read-write locks
- **Lazy Loading**: Templates loaded only when needed
- **Memory Optimization**: Shared template instances across generator instances
- **Structured Logging**: Comprehensive logging for debugging and monitoring

### Security Model

1. **Input Validation**:
   - Protocol buffer schema validation
   - Custom business rule validation
   - Cross-reference integrity checking
   - Path traversal prevention

2. **Template Security**:
   - Git repository allowlist (GitHub, GitLab, Bitbucket)
   - URL validation and normalization
   - Secure temporary directory handling
   - Automatic cleanup of cloned repositories

3. **Output Security**:
   - File path sanitization
   - Restrictive file permissions (0600 for files, 0750 for directories)
   - Sensitive value marking in Terraform outputs
   - Quote escaping for injection prevention

### Template System

The template system supports multiple sources with automatic failover:

```
Template Source Priority:
1. Local Directory (--template-dir)
2. Git Repository (--template-repo)
3. Built-in Templates (default)

Template Functions Available:
• regionToString()      - Convert enums to GCP strings
• machineTypeToString() - Machine type conversions
• quote()              - Safe string quoting
• indent()             - Text formatting
• unescapeNewlines()   - Script processing
```

#### Dependency Management

Custoodian automatically handles Terraform resource dependencies to prevent race conditions and timing issues:

**Project API Dependencies**: Networking and compute resources automatically depend on Google Cloud APIs being enabled:
```hcl
resource "google_compute_network" "vpc" {
  # Automatically waits for APIs
  depends_on = [google_project_service.api_0]
}
```

**Network Dependencies**: Compute resources automatically depend on networking resources:
```hcl
resource "google_compute_instance" "vm" {
  # Automatically references and depends on network
  depends_on = [
    google_compute_network.vpc,
    google_project_service.api_0
  ]
}
```

**Resource References**: Built-in templates automatically use proper Terraform resource references instead of string names where appropriate:
```hcl
# Firewall rules reference VPC resources
resource "google_compute_firewall" "rule" {
  network = google_compute_network.vpc.name  # Not just "vpc"
}
```

This ensures that:
- Resources are created in the correct order
- API services are enabled before dependent resources
- Cross-resource references use proper Terraform syntax
- Manual dependency management is not required

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Protocol Buffers compiler (`protoc`)
- Buf CLI tool
- Git (for Git repository template loading)

### Setup

```bash
git clone https://github.com/custoodian/custoodian.git
cd custoodian

# Install dependencies
make deps

# Generate protobuf code
make proto

# Build with optimizations
make build

# Run comprehensive tests
make test

# Code quality checks
make check

# Development build with debug info
go build -o bin/custoodian-dev -ldflags "-X main.version=dev" ./cmd/custoodian
```

### Project Structure

```
custoodian/
├── cmd/custoodian/          # CLI main package and entry point
├── internal/                # Internal packages (not for external use)
│   ├── cmd/                # CLI command implementations
│   │   ├── generate.go     # Terraform generation command
│   │   ├── validate.go     # Configuration validation command
│   │   ├── schema.go       # Schema export command
│   │   └── utils.go        # Shared utilities with security features
│   ├── generator/          # Core Terraform generation engine
│   │   ├── generator.go    # Main generation logic with caching
│   │   └── helpers.go      # Template functions and utilities
│   ├── templates/          # Template loading and management
│   │   ├── builtin.go      # Embedded templates for all GCP resources
│   │   └── loader.go       # Multi-source template loading with security
│   └── validator/          # Configuration validation engine
│       ├── validator.go    # Comprehensive validation rules
│       └── validator_test.go # Validation test suite
├── pkg/config/             # Generated protobuf Go code (public API)
├── proto/custoodian/        # Protocol buffer schema definitions
│   ├── config.proto        # Main configuration schema
│   └── enums.proto         # GCP resource enumerations
├── examples/               # Example configurations and documentation
│   ├── simple.textproto    # Basic web application setup
│   └── advanced.textproto  # Enterprise-grade configuration
├── templates/gcp/          # Reference templates for customization
└── .github/workflows/      # CI/CD automation
    ├── ci.yml             # Continuous integration
    ├── release.yml        # Release automation
    └── security.yml       # Security scanning
```

### Performance Profiling

Enable performance profiling for development:

```bash
# Build with profiling
go build -tags profile -o bin/custoodian-profile ./cmd/custoodian

# Generate with CPU profiling
./bin/custoodian-profile generate config.textproto -cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Debugging Template Issues

```bash
# Enable verbose logging
CUSTOODIAN_LOG_LEVEL=debug ./bin/custoodian generate config.textproto

# Test custom templates
custoodian generate config.textproto --template-dir ./debug-templates --dry-run

# Validate template syntax
custoodian generate --template-dir ./templates --dry-run minimal.textproto
```

### Adding New GCP Resources

1. **Add to Protocol Buffer schema** (`proto/custoodian/config.proto`):
   ```protobuf
   message NewResource {
     string name = 1 [(buf.validate.field).string.min_len = 1];
     // Add other fields with validation
   }
   ```

2. **Add to enums** if needed (`proto/custoodian/enums.proto`):
   ```protobuf
   enum NewResourceType {
     NEW_RESOURCE_TYPE_UNSPECIFIED = 0;
     NEW_RESOURCE_TYPE_STANDARD = 1;
   }
   ```

3. **Update generator** (`internal/generator/generator.go`):
   ```go
   func (g *Generator) generateNewResource(resource *config.NewResource) (string, error) {
     var output strings.Builder
     err := g.templates.ExecuteTemplate(&output, "new_resource.tf", resource)
     return output.String(), err
   }
   ```

4. **Add template** (`internal/templates/builtin.go`):
   ```go
   const newResourceTemplate = `
   resource "google_new_resource" "{{ .Name }}" {
     name = {{ quote .Name }}
     # Add other Terraform resource configuration
   }
   `
   ```

5. **Add validation** (`internal/validator/validator.go`):
   ```go
   func validateNewResource(resource *config.NewResource) error {
     // Add custom validation logic
     return nil
   }
   ```

6. **Regenerate protobuf code**:
   ```bash
   make proto
   ```

## 🤝 Contributing

We welcome contributions!
 
### Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Provide minimal reproducible examples
- Include custoodian version and environment details

### Pull Requests

- Fork the repository and create a feature branch
- Add tests for new functionality
- Ensure all tests pass: `make test`
- Update documentation as needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by Google's internal "latchkey" tool
- Built with [Protocol Buffers](https://protobuf.dev/) and [buf](https://buf.build/)
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework
- Terraform generation powered by Go's [text/template](https://pkg.go.dev/text/template)
