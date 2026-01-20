---
page_title: "agentlink_application Resource - AgentLink"
subcategory: ""
description: |-
  Manages a Frontegg application that serves as a container for MCP configurations and tools.
---

# agentlink_application (Resource)

Manages a Frontegg application that serves as a container for MCP configurations and tools.

## Example Usage

```terraform
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

## Schema

### Required

- `name` (String) Application name.
- `app_url` (String) Application URL.
- `login_url` (String) Login/OAuth URL.

### Optional

- `type` (String) Application type. Valid values: `web`, `mobile-ios`, `mobile-android`, `agent`, `other`. Defaults to `agent`.
- `access_type` (String) Access type. Valid values: `FREE_ACCESS`, `MANAGED_ACCESS`. Defaults to `FREE_ACCESS`.
- `allow_dcr` (Boolean) Enable Dynamic Client Registration. Defaults to `true`.
- `description` (String) Application description.
- `is_active` (Boolean) Whether the application is active. Defaults to `true`.
- `is_default` (Boolean) Whether this is the default application. Defaults to `false`.
- `logo_url` (String) Application logo URL.
- `frontend_stack` (String) Frontend framework. Valid values: `react`, `angular`, `vue`, `nextjs`, `other`. Defaults to `react`.

### Read-Only

- `id` (String) The application ID.
- `vendor_id` (String) The vendor ID.
- `app_host` (String) The application host (computed by Frontegg).

## Import

Import is supported using the application ID:

```shell
terraform import agentlink_application.main <application_id>
```
