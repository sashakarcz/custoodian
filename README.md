# Custoodian - GCP Terraform Generator

[![Build Status](https://github.com/custoodian/custoodian/workflows/CI/badge.svg)](https://github.com/custoodian/custoodian/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/custoodian/custoodian)](https://goreportcard.com/report/github.com/custoodian/custoodian)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Custoodian is a tool that generates Terraform code from Protocol Buffer text configurations for Google Cloud Platform resources. It provides type-safe infrastructure configuration with comprehensive validation and supports custom template systems.

Custoodian leverages Protocol Buffers for strong typing and validation, catching configuration errors before Terraform runs.

## ✨ Features

- **Protocol Buffer-based Configuration**: Type-safe infrastructure definitions with compile-time validation
- **Comprehensive GCP Support**: Full coverage of GCP resources including compute, networking, storage, and IAM
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

## 📖 Documentation

### Protocol Buffer Schema

Custoodian uses Protocol Buffers to define infrastructure configurations. The main message types include:

- `Project`: GCP project configuration, APIs, billing
- `Networking`: VPCs, subnets, firewall rules, NAT gateways, reserved IPs
- `Compute`: Instance templates, managed instance groups, individual instances
- `LoadBalancer`: HTTP/HTTPS/TCP load balancers with health checks
- `Iam`: Service accounts, role bindings, custom roles
- `Storage`: Cloud Storage buckets with lifecycle policies

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
├── variables.tf
└── outputs.tf
```

#### Git Repository Templates

```bash
custoodian generate config.textproto --template-repo github.com/myorg/gcp-templates
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
