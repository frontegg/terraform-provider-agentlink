# Terraform Provider for Frontegg AgentLink

A Terraform provider for managing Frontegg AgentLink resources, including applications and MCP (Model Context Protocol) configuration sources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)

## Installation

### From Source

```bash
make install
```

This will build the provider and install it to `~/.terraform.d/plugins/`.

## Authentication

The provider requires Frontegg API credentials. You can obtain these from the Frontegg Portal under **Settings > API Keys**.

Credentials can be provided via:

1. **Provider configuration:**
   ```hcl
   provider "agentlink" {
     client_id = "your-client-id"
     secret    = "your-secret"
     region    = "eu"
   }
   ```

2. **Environment variables:**
   ```bash
   export FRONTEGG_CLIENT_ID="your-client-id"
   export FRONTEGG_SECRET="your-secret"
   export FRONTEGG_REGION="eu"
   ```

## Configuration

### Provider Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `client_id` | Frontegg API client ID | Yes | - |
| `secret` | Frontegg API secret | Yes | - |
| `region` | Frontegg region (`stg`, `eu`, `us`, `au`, `ca`, `uk`) | No | `eu` |
| `base_url` | Override the API base URL (takes precedence over `region`) | No | - |
| `application_name` | Name of the application to use/create | No | - |
| `sources` | List of MCP configuration sources to create | No | `[]` |

### Source Configuration

Each source in the `sources` list supports:

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Name of the source | Yes | - |
| `type` | Source type: `REST`, `GRAPHQL`, `MOCK`, `MCP_PROXY`, `FRONTEGG`, `CUSTOM_INTEGRATION` | Yes | - |
| `source_url` | URL of the source (must be HTTPS) | Yes | - |
| `api_timeout` | API timeout in milliseconds (500-5000) | No | `3000` |
| `schema_file` | Path to OpenAPI/GraphQL schema file to import | No | - |

## Example Usage

```hcl
terraform {
  required_providers {
    agentlink = {
      source = "frontegg/agentlink"
    }
  }
}

provider "agentlink" {
  client_id        = var.client_id
  secret           = var.secret
  region           = "eu"
  application_name = "My Application"

  sources = [
    {
      name        = "My REST API"
      type        = "REST"
      source_url  = "https://api.example.com"
      api_timeout = 3000
      schema_file = "./openapi.json"
    }
  ]
}

# Get the current application info
data "agentlink_application" "current" {}

output "application_id" {
  value = data.agentlink_application.current.id
}

output "application_name" {
  value = data.agentlink_application.current.name
}
```

## Data Sources

### agentlink_application

Retrieves information about the currently configured application.

#### Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | The application ID |
| `name` | The application name |

## Development

### Building

```bash
make build
```

### Installing Locally

```bash
make install
```

### Running Tests

```bash
make test        # Unit tests
make testacc     # Acceptance tests
```

### Testing with Terraform

1. Copy the example tfvars file:
   ```bash
   cp examples/provider/terraform.tfvars.example examples/provider/terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your credentials

3. Run:
   ```bash
   make test-plan   # Preview changes
   make test-apply  # Apply changes
   ```

## License

See [LICENSE](LICENSE) for details.
