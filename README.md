# Custodian - GCP Terraform Generator

[![Build Status](https://github.com/custodian/custodian/workflows/CI/badge.svg)](https://github.com/custodian/custodian/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/custodian/custodian)](https://goreportcard.com/report/github.com/custodian/custodian)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Custodian is a tool that generates Terraform code from Protocol Buffer text configurations for Google Cloud Platform resources. It provides type-safe infrastructure configuration with comprehensive validation and supports custom template systems.

Custodian leverages Protocol Buffers for strong typing and validation, catching configuration errors before Terraform runs.

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
curl -L -o /usr/local/bin/custodian https://github.com/custodian/custodian/releases/latest/download/custodian-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')
chmod +x /usr/local/bin/custodian
```

#### Build from Source

```bash
git clone https://github.com/custodian/custodian.git
cd custodian
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
  apis: [API_COMPUTE, API_STORAGE, API_IAM]
}

networking {
  reserved_ips {
    name: "web-lb-ip"
    type: GLOBAL
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
    machine_type: MACHINE_E2_MEDIUM
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
  type: LB_HTTP
  ip: "web-lb-ip"
  backend: "web-group"
}
```

2. **Validate the configuration**:

```bash
custodian validate config.textproto
```

3. **Generate Terraform code**:

```bash
custodian generate config.textproto --output ./terraform
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

Custodian uses Protocol Buffers to define infrastructure configurations. The main message types include:

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
  MACHINE_E2_MICRO = 1;
  MACHINE_E2_MEDIUM = 3;
  MACHINE_N1_STANDARD_4 = 10;
  // ... more machine types
}
```

### CLI Commands

#### Generate Terraform Code

```bash
# Basic generation
custodian generate config.textproto

# With custom output directory
custodian generate config.textproto --output ./infrastructure

# Using custom templates
custodian generate config.textproto --template-dir ./custom-templates

# Using templates from Git repository
custodian generate config.textproto --template-repo github.com/org/templates

# Dry run (show what would be generated)
custodian generate config.textproto --dry-run
```

#### Validate Configuration

```bash
# Validate syntax and constraints
custodian validate config.textproto
```

#### Display Schema

```bash
# Show Protocol Buffer schema
custodian schema

# Export schema to directory
custodian schema --output ./schema
```

### Custom Templates

Custodian supports custom Terraform templates for organizations that need specific patterns:

#### Local Template Directory

```bash
custodian generate config.textproto --template-dir ./my-templates
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
custodian generate config.textproto --template-repo github.com/myorg/gcp-templates
```

### GitHub Action

Use Custodian in your CI/CD pipeline:

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
        uses: custodian/custodian@v1
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

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Protocol Buffers compiler (`protoc`)
- Buf CLI tool

### Setup

```bash
git clone https://github.com/custodian/custodian.git
cd custodian

# Install dependencies
make deps

# Generate protobuf code
make proto

# Build
make build

# Run tests
make test

# Format and lint
make check
```

### Project Structure

```
custodian/
├── cmd/custodian/          # CLI main package
├── internal/
│   ├── cmd/                # CLI commands
│   ├── generator/          # Terraform generation
│   ├── templates/          # Template system
│   └── validator/          # Configuration validation
├── pkg/config/             # Generated protobuf code
├── proto/custodian/        # Protocol buffer schemas
├── examples/               # Example configurations
├── templates/gcp/          # Built-in templates
└── .github/workflows/      # CI/CD workflows
```

### Adding New GCP Resources

1. **Add to Protocol Buffer schema** (`proto/custodian/config.proto`):
   ```protobuf
   message NewResource {
     string name = 1 [(buf.validate.field).string.min_len = 1];
     // Add other fields with validation
   }
   ```

2. **Add to enums** if needed (`proto/custodian/enums.proto`):
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
- Include custodian version and environment details

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
