---
page_title: "agentlink_tools_import Resource - AgentLink"
subcategory: ""
description: |-
  Imports tools from OpenAPI or GraphQL schema files.
---

# agentlink_tools_import (Resource)

Imports tools from OpenAPI (Swagger) or GraphQL schema files. Tools are automatically discovered and made available to your AI agent.

## Example Usage

```terraform
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

## Schema

### Required

- `application_id` (String) Application ID. Changing this forces a new resource to be created.
- `source_id` (String) Source ID to associate tools with. Changing this forces a new resource to be created.
- `schema_file` (String) Path to OpenAPI (JSON/YAML) or GraphQL schema file.
- `schema_type` (String) Schema type. Valid values: `openapi`, `graphql`. Changing this forces a new resource to be created.

### Read-Only

- `id` (String) Composite ID (app_id:source_id).
- `schema_hash` (String) SHA256 hash of schema content. Changes trigger reimport.
- `tools_count` (Number) Number of tools imported.
