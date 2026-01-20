---
page_title: "agentlink_mcp_configuration Resource - AgentLink"
subcategory: ""
description: |-
  Configures Model Context Protocol (MCP) settings for an application.
---

# agentlink_mcp_configuration (Resource)

Configures Model Context Protocol (MCP) settings for an application. MCP is the standard protocol for AI model interactions with external tools and APIs.

## Example Usage

```terraform
resource "agentlink_mcp_configuration" "main" {
  application_id = agentlink_application.main.id
  base_url       = "https://api.example.com"
  api_timeout    = 5000
}
```

## Schema

### Required

- `application_id` (String) Application ID. Changing this forces a new resource to be created.
- `base_url` (String) Base URL for API tool calls.

### Optional

- `api_timeout` (Number) API timeout in milliseconds. Defaults to `5000`.

### Read-Only

- `id` (String) The resource ID.

## Import

Import is supported using the application ID:

```shell
terraform import agentlink_mcp_configuration.main <application_id>
```
