# Terraform Provider for Frontegg AgentLink

A Terraform provider for managing Frontegg AgentLink resources, including applications, MCP configurations, sources, tools, and policies.

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

## Provider Configuration

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `client_id` | Frontegg API client ID | Yes | - |
| `secret` | Frontegg API secret | Yes | - |
| `region` | Frontegg region (`stg`, `eu`, `us`, `au`, `ca`, `uk`) | No | `eu` |
| `base_url` | Override the API base URL (takes precedence over `region`) | No | - |

## Resources

### agentlink_application

Manages a Frontegg application.

```hcl
resource "agentlink_application" "main" {
  name        = "My MCP Server"
  app_url     = "https://app.example.com"
  login_url   = "https://app.example.com/oauth"
  type        = "agent"
  allow_dcr   = true
  description = "My MCP server application"
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Application name | Yes | - |
| `app_url` | Application URL | Yes | - |
| `login_url` | Login URL | Yes | - |
| `type` | Application type (`web`, `mobile-ios`, `mobile-android`, `agent`, `other`) | No | `agent` |
| `access_type` | Access type (`FREE_ACCESS`, `MANAGED_ACCESS`) | No | `FREE_ACCESS` |
| `allow_dcr` | Enable Dynamic Client Registration | No | `true` |
| `description` | Application description | No | - |
| `is_active` | Whether the application is active | No | `true` |
| `is_default` | Whether this is the default application | No | `false` |

#### Attributes

- `id` - The application ID
- `vendor_id` - The vendor ID
- `app_host` - The application host

---

### agentlink_mcp_configuration

Manages MCP (Model Context Protocol) configuration for an application.

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
| `application_id` | Application ID | Yes | - |
| `base_url` | API base URL for tools | Yes | - |
| `api_timeout` | API timeout in milliseconds | No | `5000` |

---

### agentlink_source

Manages an MCP configuration source.

```hcl
resource "agentlink_source" "rest_api" {
  application_id = agentlink_application.main.id
  name           = "My REST API"
  type           = "REST"
  source_url     = "https://api.example.com"
  api_timeout    = 3000
  enabled        = true
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `application_id` | Application ID | Yes | - |
| `name` | Source name | Yes | - |
| `type` | Source type (`REST`, `GRAPHQL`, `MOCK`, `MCP_PROXY`, `FRONTEGG`, `CUSTOM_INTEGRATION`) | Yes | - |
| `source_url` | Source URL (must be HTTPS) | Yes | - |
| `api_timeout` | API timeout in milliseconds (500-5000) | No | `3000` |
| `enabled` | Whether the source is enabled | No | `true` |

---

### agentlink_tools_import

Imports tools from an OpenAPI or GraphQL schema.

```hcl
resource "agentlink_tools_import" "openapi" {
  application_id = agentlink_application.main.id
  source_id      = agentlink_source.rest_api.id
  schema_file    = "./openapi.json"
  schema_type    = "openapi"
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `application_id` | Application ID | Yes | - |
| `source_id` | Source ID to associate tools with | Yes | - |
| `schema_file` | Path to schema file | Yes | - |
| `schema_type` | Schema type (`openapi`, `graphql`) | Yes | - |

#### Attributes

- `schema_hash` - SHA256 hash of the schema content (triggers reimport on change)
- `tools_count` - Number of tools imported

---

### agentlink_rbac_policy

Manages an RBAC (Role-Based Access Control) policy.

```hcl
resource "agentlink_rbac_policy" "admin_only" {
  name              = "Admin Only Tools"
  description       = "Restrict sensitive tools to admin users"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["admin", "super-admin"]
  internal_tool_ids = ["tool-id-1", "tool-id-2"]
}
```

#### Arguments

| Argument | Description | Required | Default |
|----------|-------------|----------|---------|
| `name` | Policy name | Yes | - |
| `description` | Policy description | No | - |
| `enabled` | Whether the policy is enabled | Yes | - |
| `type` | RBAC type (`RBAC_ROLES`, `RBAC_PERMISSIONS`) | Yes | - |
| `keys` | List of role or permission keys | Yes | - |
| `internal_tool_ids` | List of tool IDs (at least one required) | Yes | - |
| `app_ids` | List of application IDs | No | - |
| `tenant_id` | Tenant ID | No | - |

---

### agentlink_masking_policy

Manages a data masking policy for sensitive information protection.

```hcl
resource "agentlink_masking_policy" "pii_protection" {
  name              = "PII Data Masking"
  description       = "Mask personally identifiable information"
  enabled           = true
  internal_tool_ids = []  # Apply to all tools

  policy_configuration {
    credit_card   = true
    email_address = true
    phone_number  = true
    us_ssn        = true
    ip_address    = false
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
| `policy_configuration` | Masking configuration block | Yes | - |
| `app_ids` | List of application IDs | No | - |
| `tenant_id` | Tenant ID | No | - |

#### Policy Configuration

| Attribute | Description |
|-----------|-------------|
| `credit_card` | Mask credit card numbers |
| `email_address` | Mask email addresses |
| `phone_number` | Mask phone numbers |
| `ip_address` | Mask IP addresses |
| `us_ssn` | Mask US Social Security Numbers |
| `us_driver_license` | Mask US driver licenses |
| `us_passport` | Mask US passports |
| `us_itin` | Mask US ITINs |
| `us_bank_number` | Mask US bank numbers |
| `iban_code` | Mask IBAN codes |
| `swift_code` | Mask SWIFT codes |
| `bitcoin_address` | Mask Bitcoin addresses |
| `ethereum_address` | Mask Ethereum addresses |
| `cvv_cvc` | Mask CVV/CVC codes |
| `url` | Mask URLs |

---

### agentlink_conditional_policy

Manages a conditional policy with custom targeting rules.

```hcl
resource "agentlink_conditional_policy" "approval_required" {
  name              = "Require Approval for Deletes"
  description       = "Delete operations require manager approval"
  enabled           = true
  internal_tool_ids = []  # Apply to all tools

  targeting {
    if {
      condition {
        attribute = "tool.method"
        negate    = false
        op        = "in_list"
        value     = { list = "DELETE" }
      }
    }
    then {
      result           = "APPROVAL_REQUIRED"
      approval_flow_id = "approval-flow-uuid"
    }
  }
}
```

### agentlink_allowed_origins

Manages the allowed origins (CORS) configuration for the Frontegg vendor.

```hcl
resource "agentlink_allowed_origins" "cors" {
  allowed_origins = [
    "http://localhost:3000",
    "https://app.example.com",
    "https://staging.example.com"
  ]
}
```

#### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `allowed_origins` | List of allowed origins for CORS | Yes |

#### Attributes

- `id` - The vendor ID

---

## Data Sources

### agentlink_application

Retrieves information about the current application.

```hcl
data "agentlink_application" "current" {}

output "app_id" {
  value = data.agentlink_application.current.id
}
```

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
  client_id = var.client_id
  secret    = var.secret
  region    = "eu"
}

# Create an application
resource "agentlink_application" "main" {
  name      = "My MCP Server"
  app_url   = "https://app.example.com"
  login_url = "https://app.example.com/oauth"
  type      = "agent"
  allow_dcr = true
}

# Configure MCP
resource "agentlink_mcp_configuration" "main" {
  application_id = agentlink_application.main.id
  base_url       = "https://api.example.com"
  api_timeout    = 5000
}

# Create a source
resource "agentlink_source" "api" {
  application_id = agentlink_application.main.id
  name           = "My API"
  type           = "REST"
  source_url     = "https://api.example.com"
  enabled        = true
}

# Import tools from OpenAPI spec
resource "agentlink_tools_import" "api_tools" {
  application_id = agentlink_application.main.id
  source_id      = agentlink_source.api.id
  schema_file    = "./openapi.json"
  schema_type    = "openapi"
}

# Add RBAC policy
resource "agentlink_rbac_policy" "admin_tools" {
  name              = "Admin Only"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["admin"]
  internal_tool_ids = []
}

# Add masking policy
resource "agentlink_masking_policy" "pii" {
  name              = "PII Masking"
  enabled           = true
  internal_tool_ids = []

  policy_configuration {
    credit_card   = true
    email_address = true
    us_ssn        = true
  }
}

# Configure allowed origins (CORS)
resource "agentlink_allowed_origins" "cors" {
  allowed_origins = [
    "http://localhost:3000",
    "https://app.example.com"
  ]
}
```

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
