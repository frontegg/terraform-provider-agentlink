# Terraform Provider for Frontegg AgentLink

[![CI](https://github.com/frontegg/terraform-provider-agentlink/actions/workflows/test.yml/badge.svg)](https://github.com/frontegg/terraform-provider-agentlink/actions/workflows/test.yml)
[![Release](https://github.com/frontegg/terraform-provider-agentlink/actions/workflows/release.yml/badge.svg)](https://github.com/frontegg/terraform-provider-agentlink/actions/workflows/release.yml)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-blue.svg)](https://registry.terraform.io/providers/frontegg/agentlink/latest)

A Terraform provider for managing [Frontegg AgentLink](https://frontegg.com) resources. AgentLink enables you to build and deploy secure AI agents with enterprise-grade access control, data masking, and Model Context Protocol (MCP) integration.

## What is Frontegg AgentLink?

Frontegg AgentLink is a platform for building secure AI agents that can interact with your APIs and services. It provides:

- **Model Context Protocol (MCP) Support**: Standard protocol for AI model interactions with tools and APIs
- **Enterprise Security**: Role-based access control (RBAC), data masking, and conditional policies
- **Multi-Source Integration**: Connect REST APIs, GraphQL endpoints, and custom integrations
- **Tool Management**: Import tools from OpenAPI/GraphQL schemas with automatic discovery

This Terraform provider allows you to manage all AgentLink resources as Infrastructure as Code (IaC), enabling version control, reproducibility, and automation of your AI agent infrastructure.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Provider Configuration](#provider-configuration)
- [Resources](#resources)
  - [agentlink_application](#agentlink_application)
  - [agentlink_mcp_configuration](#agentlink_mcp_configuration)
  - [agentlink_source](#agentlink_source)
  - [agentlink_tools_import](#agentlink_tools_import)
  - [agentlink_rbac_policy](#agentlink_rbac_policy)
  - [agentlink_masking_policy](#agentlink_masking_policy)
  - [agentlink_conditional_policy](#agentlink_conditional_policy)
  - [agentlink_allowed_origins](#agentlink_allowed_origins)
- [Data Sources](#data-sources)
- [Complete Example](#complete-example)
- [Security Best Practices](#security-best-practices)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (only for building from source)
- A Frontegg account with AgentLink access

## Installation

### From Terraform Registry (Recommended)

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    agentlink = {
      source  = "frontegg/agentlink"
      version = "~> 0.1"
    }
  }
}
```

Then run:

```bash
terraform init
```

### From Source

Clone the repository and build:

```bash
git clone https://github.com/frontegg/terraform-provider-agentlink.git
cd terraform-provider-agentlink
make install
```

This installs the provider to `~/.terraform.d/plugins/`.

## Quick Start

1. **Set up authentication** using environment variables:

```bash
export FRONTEGG_CLIENT_ID="your-client-id"
export FRONTEGG_SECRET="your-secret"
```

2. **Create a basic configuration** (`main.tf`):

```hcl
terraform {
  required_providers {
    agentlink = {
      source = "frontegg/agentlink"
    }
  }
}

provider "agentlink" {}

# Create an AI agent application
resource "agentlink_application" "my_agent" {
  name      = "My AI Agent"
  app_url   = "https://agent.example.com"
  login_url = "https://agent.example.com/login"
  type      = "agent"
  allow_dcr = true
}

# Configure MCP settings
resource "agentlink_mcp_configuration" "config" {
  application_id = agentlink_application.my_agent.id
  base_url       = "https://api.example.com"
}
```

3. **Apply the configuration**:

```bash
terraform init
terraform plan
terraform apply
```

## Authentication

The provider requires Frontegg API credentials. Obtain these from the [Frontegg Portal](https://portal.frontegg.com) under **Settings > API Keys**.

### Option 1: Environment Variables (Recommended)

```bash
export FRONTEGG_CLIENT_ID="your-client-id"
export FRONTEGG_SECRET="your-secret"
export FRONTEGG_REGION="eu"  # Optional, defaults to "eu"
```

### Option 2: Provider Configuration

```hcl
provider "agentlink" {
  client_id = "your-client-id"
  secret    = "your-secret"
  region    = "eu"
}
```

### Option 3: Using Variables (for CI/CD)

```hcl
variable "frontegg_client_id" {
  type      = string
  sensitive = true
}

variable "frontegg_secret" {
  type      = string
  sensitive = true
}

provider "agentlink" {
  client_id = var.frontegg_client_id
  secret    = var.frontegg_secret
}
```

## Provider Configuration

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `client_id` | Frontegg API client ID | Yes | `FRONTEGG_CLIENT_ID` env var |
| `secret` | Frontegg API secret | Yes | `FRONTEGG_SECRET` env var |
| `region` | Frontegg region | No | `eu` |
| `base_url` | Override API base URL | No | Derived from region |

### Supported Regions

| Region | API Endpoint |
|--------|--------------|
| `stg` | https://api.stg.frontegg.com |
| `eu` | https://api.frontegg.com |
| `us` | https://api.us.frontegg.com |
| `au` | https://api.au.frontegg.com |
| `ca` | https://api.ca.frontegg.com |
| `uk` | https://api.uk.frontegg.com |

## Resources

### agentlink_application

Manages a Frontegg application that serves as a container for MCP configurations and tools.

```hcl
resource "agentlink_application" "main" {
  name        = "My MCP Server"
  app_url     = "https://app.example.com"
  login_url   = "https://app.example.com/oauth"
  type        = "agent"
  access_type = "MANAGED_ACCESS"
  allow_dcr   = true
  description = "Production AI agent application"
  is_active   = true
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Application name | Yes | - |
| `app_url` | Application URL | Yes | - |
| `login_url` | Login/OAuth URL | Yes | - |
| `type` | Application type: `web`, `mobile-ios`, `mobile-android`, `agent`, `other` | No | `agent` |
| `access_type` | Access type: `FREE_ACCESS`, `MANAGED_ACCESS` | No | `FREE_ACCESS` |
| `allow_dcr` | Enable Dynamic Client Registration | No | `true` |
| `description` | Application description | No | - |
| `is_active` | Whether the application is active | No | `true` |
| `is_default` | Whether this is the default application | No | `false` |
| `logo_url` | Application logo URL | No | - |
| `frontend_stack` | Frontend framework: `react`, `angular`, `vue`, `nextjs`, `other` | No | `react` |

#### Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | The application ID |
| `vendor_id` | The vendor ID |
| `app_host` | The application host (computed by Frontegg) |

---

### agentlink_mcp_configuration

Configures Model Context Protocol (MCP) settings for an application. MCP is the standard protocol for AI model interactions with external tools and APIs.

```hcl
resource "agentlink_mcp_configuration" "main" {
  application_id = agentlink_application.main.id
  base_url       = "https://api.example.com"
  api_timeout    = 5000
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `application_id` | Application ID (forces replacement on change) | Yes | - |
| `base_url` | Base URL for API tool calls | Yes | - |
| `api_timeout` | API timeout in milliseconds | No | `5000` |

---

### agentlink_source

Manages an MCP configuration source. Sources define where your AI agent can discover and execute tools.

```hcl
resource "agentlink_source" "rest_api" {
  application_id = agentlink_application.main.id
  name           = "Customer API"
  type           = "REST"
  source_url     = "https://api.example.com"
  api_timeout    = 3000
  enabled        = true
}

resource "agentlink_source" "graphql_api" {
  application_id = agentlink_application.main.id
  name           = "GraphQL Gateway"
  type           = "GRAPHQL"
  source_url     = "https://graphql.example.com"
  enabled        = true
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `application_id` | Application ID (forces replacement) | Yes | - |
| `name` | Source name | Yes | - |
| `type` | Source type (see below) | Yes | - |
| `source_url` | Source URL (must be HTTPS) | Yes | - |
| `api_timeout` | API timeout in milliseconds (500-5000) | No | `3000` |
| `enabled` | Whether the source is enabled | No | `true` |

#### Source Types

| Type | Description |
|------|-------------|
| `REST` | RESTful API endpoints |
| `GRAPHQL` | GraphQL API endpoints |
| `MOCK` | Mock/testing endpoints |
| `MCP_PROXY` | MCP proxy server |
| `FRONTEGG` | Frontegg internal APIs |
| `CUSTOM_INTEGRATION` | Custom integration |

---

### agentlink_tools_import

Imports tools from OpenAPI (Swagger) or GraphQL schema files. Tools are automatically discovered and made available to your AI agent.

```hcl
resource "agentlink_tools_import" "openapi_tools" {
  application_id = agentlink_application.main.id
  source_id      = agentlink_source.rest_api.id
  schema_file    = "${path.module}/schemas/openapi.json"
  schema_type    = "openapi"
}

resource "agentlink_tools_import" "graphql_tools" {
  application_id = agentlink_application.main.id
  source_id      = agentlink_source.graphql_api.id
  schema_file    = "${path.module}/schemas/schema.graphql"
  schema_type    = "graphql"
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `application_id` | Application ID (forces replacement) | Yes | - |
| `source_id` | Source ID to associate tools with (forces replacement) | Yes | - |
| `schema_file` | Path to OpenAPI (JSON/YAML) or GraphQL schema file | Yes | - |
| `schema_type` | Schema type: `openapi` or `graphql` (forces replacement) | Yes | - |

#### Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | Composite ID (app_id:source_id) |
| `schema_hash` | SHA256 hash of schema content (triggers reimport on change) |
| `tools_count` | Number of tools imported |

---

### agentlink_rbac_policy

Manages Role-Based Access Control (RBAC) policies. Restrict which tools can be accessed based on user roles or permissions.

```hcl
# Restrict by roles
resource "agentlink_rbac_policy" "admin_tools" {
  name              = "Admin Only Tools"
  description       = "Restrict sensitive tools to admin users"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["admin", "super-admin"]
  internal_tool_ids = ["tool-id-1", "tool-id-2"]
  app_ids           = [agentlink_application.main.id]
}

# Restrict by permissions
resource "agentlink_rbac_policy" "write_permission" {
  name              = "Write Permission Required"
  description       = "Tools that modify data require write permission"
  enabled           = true
  type              = "RBAC_PERMISSIONS"
  keys              = ["data:write", "data:admin"]
  internal_tool_ids = []  # Empty = apply to all tools
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Policy name | Yes | - |
| `description` | Policy description | No | - |
| `enabled` | Whether the policy is enabled | Yes | - |
| `type` | RBAC type: `RBAC_ROLES` or `RBAC_PERMISSIONS` (forces replacement) | Yes | - |
| `keys` | List of role or permission keys (at least one) | Yes | - |
| `internal_tool_ids` | List of tool IDs (at least one, or empty for all) | Yes | - |
| `app_ids` | List of application IDs to apply policy to | No | - |
| `tenant_id` | Tenant ID for multi-tenant scenarios | No | - |

---

### agentlink_masking_policy

Manages data masking policies for protecting sensitive information like PII, financial data, and crypto addresses. Masking policies automatically redact sensitive data in tool responses.

```hcl
resource "agentlink_masking_policy" "pii_protection" {
  name              = "PII Data Masking"
  description       = "Mask personally identifiable information in all responses"
  enabled           = true
  internal_tool_ids = []  # Apply to all tools

  policy_configuration {
    # Personal Information
    email_address = true
    phone_number  = true
    ip_address    = true

    # US-specific PII
    us_ssn            = true
    us_driver_license = true
    us_passport       = true
    us_itin           = true

    # Financial Information
    credit_card    = true
    cvv_cvc        = true
    us_bank_number = true
    iban_code      = true
    swift_code     = true

    # Cryptocurrency
    bitcoin_address  = true
    ethereum_address = true

    # Other
    url = false
  }
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Policy name | Yes | - |
| `description` | Policy description | No | - |
| `enabled` | Whether the policy is enabled | Yes | - |
| `internal_tool_ids` | List of tool IDs (empty = all tools) | Yes | - |
| `policy_configuration` | Masking configuration block (see below) | Yes | - |
| `app_ids` | List of application IDs | No | - |
| `tenant_id` | Tenant ID | No | - |

#### Policy Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `credit_card` | Mask credit card numbers | `false` |
| `email_address` | Mask email addresses | `false` |
| `phone_number` | Mask phone numbers | `false` |
| `ip_address` | Mask IP addresses | `false` |
| `us_ssn` | Mask US Social Security Numbers | `false` |
| `us_driver_license` | Mask US driver licenses | `false` |
| `us_passport` | Mask US passport numbers | `false` |
| `us_itin` | Mask US Individual Taxpayer Identification Numbers | `false` |
| `us_bank_number` | Mask US bank account numbers | `false` |
| `iban_code` | Mask International Bank Account Numbers | `false` |
| `swift_code` | Mask SWIFT/BIC codes | `false` |
| `bitcoin_address` | Mask Bitcoin wallet addresses | `false` |
| `ethereum_address` | Mask Ethereum wallet addresses | `false` |
| `cvv_cvc` | Mask credit card CVV/CVC codes | `false` |
| `url` | Mask URLs | `false` |

---

### agentlink_conditional_policy

Manages conditional policies with advanced targeting rules. Use these for complex access control scenarios, approval workflows, and context-aware security.

```hcl
# Require approval for destructive operations
resource "agentlink_conditional_policy" "delete_approval" {
  name              = "Delete Operations Require Approval"
  description       = "All DELETE operations must be approved by a manager"
  enabled           = true
  internal_tool_ids = []

  targeting {
    if {
      condition {
        attribute = "tool.method"
        negate    = false
        op        = "equals"
        value     = { string = "DELETE" }
      }
    }
    then {
      result           = "APPROVAL_REQUIRED"
      approval_flow_id = "manager-approval-flow-id"
    }
  }
}

# Deny access outside business hours
resource "agentlink_conditional_policy" "business_hours" {
  name              = "Business Hours Only"
  description       = "Sensitive tools only available during business hours"
  enabled           = true
  internal_tool_ids = ["sensitive-tool-1", "sensitive-tool-2"]

  targeting {
    if {
      condition {
        attribute = "request.hour"
        negate    = false
        op        = "not_in_range"
        value     = { range_start = "9", range_end = "17" }
      }
    }
    then {
      result = "DENY"
    }
  }
}

# Allow specific users
resource "agentlink_conditional_policy" "allowed_users" {
  name              = "Allowed Users"
  enabled           = true
  internal_tool_ids = []

  targeting {
    if {
      condition {
        attribute = "user.email"
        negate    = false
        op        = "in_list"
        value     = { list = "admin@example.com,security@example.com" }
      }
    }
    then {
      result = "ALLOW"
    }
  }
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Policy name | Yes | - |
| `description` | Policy description | No | - |
| `enabled` | Whether the policy is enabled | Yes | - |
| `internal_tool_ids` | List of tool IDs (empty = all tools) | Yes | - |
| `targeting` | Targeting rules block | No | - |
| `app_ids` | List of application IDs | No | - |
| `tenant_id` | Tenant ID | No | - |
| `metadata` | Additional metadata map | No | - |

#### Targeting Block

| Block | Description |
|-------|-------------|
| `if` | Contains `condition` blocks with the rule conditions |
| `then` | Contains `result` (ALLOW, DENY, APPROVAL_REQUIRED) and optional `approval_flow_id` |

#### Condition Attributes

| Field | Description |
|-------|-------------|
| `attribute` | The attribute to evaluate (e.g., `tool.method`, `user.email`, `request.hour`) |
| `negate` | Whether to negate the condition |
| `op` | Operator: `equals`, `not_equals`, `in_list`, `not_in_list`, `in_range`, `not_in_range` |
| `value` | Value map with keys like `string`, `list`, `range_start`, `range_end` |

---

### agentlink_allowed_origins

Manages CORS (Cross-Origin Resource Sharing) configuration for your Frontegg vendor.

```hcl
resource "agentlink_allowed_origins" "cors" {
  allowed_origins = [
    "http://localhost:3000",
    "http://localhost:8080",
    "https://app.example.com",
    "https://staging.example.com",
    "https://admin.example.com"
  ]
}
```

#### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `allowed_origins` | Set of allowed origin URLs for CORS | Yes |

#### Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | The vendor ID |

---

## Data Sources

### agentlink_application

Retrieves information about the current application configured for the provider.

```hcl
data "agentlink_application" "current" {}

output "application_info" {
  value = {
    id   = data.agentlink_application.current.id
    name = data.agentlink_application.current.name
  }
}
```

---

## Complete Example

Here's a comprehensive example that sets up a complete AI agent infrastructure:

```hcl
terraform {
  required_providers {
    agentlink = {
      source  = "frontegg/agentlink"
      version = "~> 0.1"
    }
  }
}

# Variables
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

# Provider configuration
provider "agentlink" {
  region = "eu"
}

# Create the main application
resource "agentlink_application" "agent" {
  name        = "AI Assistant - ${var.environment}"
  app_url     = "https://assistant.example.com"
  login_url   = "https://assistant.example.com/auth"
  type        = "agent"
  access_type = "MANAGED_ACCESS"
  allow_dcr   = true
  description = "Production AI assistant with MCP integration"
}

# Configure MCP settings
resource "agentlink_mcp_configuration" "config" {
  application_id = agentlink_application.agent.id
  base_url       = "https://api.example.com"
  api_timeout    = 10000
}

# REST API source
resource "agentlink_source" "customer_api" {
  application_id = agentlink_application.agent.id
  name           = "Customer Management API"
  type           = "REST"
  source_url     = "https://api.example.com/customers"
  api_timeout    = 5000
  enabled        = true
}

# GraphQL source
resource "agentlink_source" "analytics_api" {
  application_id = agentlink_application.agent.id
  name           = "Analytics GraphQL API"
  type           = "GRAPHQL"
  source_url     = "https://analytics.example.com/graphql"
  enabled        = true
}

# Import tools from OpenAPI
resource "agentlink_tools_import" "customer_tools" {
  application_id = agentlink_application.agent.id
  source_id      = agentlink_source.customer_api.id
  schema_file    = "${path.module}/schemas/customer-api.json"
  schema_type    = "openapi"
}

# Import tools from GraphQL
resource "agentlink_tools_import" "analytics_tools" {
  application_id = agentlink_application.agent.id
  source_id      = agentlink_source.analytics_api.id
  schema_file    = "${path.module}/schemas/analytics.graphql"
  schema_type    = "graphql"
}

# RBAC: Admin-only tools
resource "agentlink_rbac_policy" "admin_only" {
  name              = "Admin Only Tools"
  description       = "Restrict administrative tools to admin role"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["admin", "super-admin"]
  internal_tool_ids = []
  app_ids           = [agentlink_application.agent.id]
}

# RBAC: Read permission for analytics
resource "agentlink_rbac_policy" "analytics_read" {
  name              = "Analytics Read Access"
  description       = "Require analytics:read permission for analytics tools"
  enabled           = true
  type              = "RBAC_PERMISSIONS"
  keys              = ["analytics:read", "analytics:admin"]
  internal_tool_ids = []
  app_ids           = [agentlink_application.agent.id]
}

# Masking: Protect PII data
resource "agentlink_masking_policy" "pii" {
  name              = "PII Protection"
  description       = "Mask all personally identifiable information"
  enabled           = true
  internal_tool_ids = []
  app_ids           = [agentlink_application.agent.id]

  policy_configuration {
    email_address     = true
    phone_number      = true
    credit_card       = true
    us_ssn            = true
    us_driver_license = true
    ip_address        = true
  }
}

# Conditional: Require approval for deletions
resource "agentlink_conditional_policy" "delete_approval" {
  name              = "Deletion Approval Required"
  description       = "All delete operations require manager approval"
  enabled           = true
  internal_tool_ids = []
  app_ids           = [agentlink_application.agent.id]

  targeting {
    if {
      condition {
        attribute = "tool.method"
        op        = "equals"
        negate    = false
        value     = { string = "DELETE" }
      }
    }
    then {
      result = "APPROVAL_REQUIRED"
    }
  }
}

# CORS configuration
resource "agentlink_allowed_origins" "cors" {
  allowed_origins = [
    "https://assistant.example.com",
    "https://admin.example.com"
  ]
}

# Outputs
output "application_id" {
  description = "The application ID"
  value       = agentlink_application.agent.id
}

output "application_host" {
  description = "The application host"
  value       = agentlink_application.agent.app_host
}

output "tools_imported" {
  description = "Number of tools imported"
  value = {
    customer  = agentlink_tools_import.customer_tools.tools_count
    analytics = agentlink_tools_import.analytics_tools.tools_count
  }
}
```

---

## Security Best Practices

### 1. Use Environment Variables for Credentials

Never commit credentials to version control:

```bash
export FRONTEGG_CLIENT_ID="your-client-id"
export FRONTEGG_SECRET="your-secret"
```

### 2. Enable Data Masking

Always enable masking for sensitive data:

```hcl
resource "agentlink_masking_policy" "default" {
  name              = "Default PII Masking"
  enabled           = true
  internal_tool_ids = []

  policy_configuration {
    email_address = true
    phone_number  = true
    credit_card   = true
    us_ssn        = true
  }
}
```

### 3. Implement Least Privilege with RBAC

Restrict tool access by roles:

```hcl
resource "agentlink_rbac_policy" "least_privilege" {
  name              = "Principle of Least Privilege"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["specific-role"]
  internal_tool_ids = ["specific-tool-id"]
}
```

### 4. Require Approval for Destructive Operations

```hcl
resource "agentlink_conditional_policy" "destructive_ops" {
  name              = "Approval for Destructive Operations"
  enabled           = true
  internal_tool_ids = []

  targeting {
    if {
      condition {
        attribute = "tool.method"
        op        = "in_list"
        negate    = false
        value     = { list = "DELETE,PUT,PATCH" }
      }
    }
    then {
      result = "APPROVAL_REQUIRED"
    }
  }
}
```

### 5. Use State Encryption

When using remote state, ensure it's encrypted:

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state"
    key            = "agentlink/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

---

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
make testacc     # Acceptance tests (requires credentials)
```

### Code Quality

```bash
make fmt         # Format code
make vet         # Run go vet
make tidy        # Tidy dependencies
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

---

## Troubleshooting

### Authentication Errors

**Error:** `401 Unauthorized`

- Verify your `client_id` and `secret` are correct
- Check that credentials have appropriate permissions
- Ensure you're using the correct `region`

### Resource Not Found

**Error:** `404 Not Found`

- Verify the resource ID exists
- Check that you're authenticated to the correct environment
- Ensure the resource wasn't deleted outside of Terraform

### HTTPS Required for Source URLs

**Error:** `source_url must use HTTPS`

- All source URLs must use HTTPS for security
- Update your `source_url` to use `https://` prefix

### Schema Import Failures

**Error:** `Failed to import tools from schema`

- Verify the schema file path is correct
- Ensure the schema is valid OpenAPI/GraphQL
- Check that `schema_type` matches the file content

### State Drift

If resources are modified outside Terraform:

```bash
terraform refresh  # Update state with actual resource state
terraform plan     # Review differences
terraform apply    # Apply to restore desired state
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation for user-facing changes
- Run `make fmt` and `make vet` before committing

---

## License

See [LICENSE](LICENSE) for details.

---

## Resources

- [Frontegg Documentation](https://docs.frontegg.com)
- [Terraform Documentation](https://www.terraform.io/docs)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io)
- [Report Issues](https://github.com/frontegg/terraform-provider-agentlink/issues)
