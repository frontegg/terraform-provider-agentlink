---
page_title: "agentlink_source Resource - AgentLink"
subcategory: ""
description: |-
  Manages an MCP configuration source for tool discovery and execution.
---

# agentlink_source (Resource)

Manages an MCP configuration source. Sources define where your AI agent can discover and execute tools.

## Example Usage

```terraform
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

## Schema

### Required

- `application_id` (String) Application ID. Changing this forces a new resource to be created.
- `name` (String) Source name.
- `type` (String) Source type. Valid values: `REST`, `GRAPHQL`, `MOCK`, `MCP_PROXY`, `FRONTEGG`, `CUSTOM_INTEGRATION`.
- `source_url` (String) Source URL. Must use HTTPS.

### Optional

- `api_timeout` (Number) API timeout in milliseconds (500-5000). Defaults to `3000`.
- `enabled` (Boolean) Whether the source is enabled. Defaults to `true`.

### Read-Only

- `id` (String) The source ID.

## Import

Import is supported using the format `application_id:source_id`:

```shell
terraform import agentlink_source.rest_api <application_id>:<source_id>
```
