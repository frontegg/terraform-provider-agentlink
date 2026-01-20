---
page_title: "agentlink_allowed_origins Resource - AgentLink"
subcategory: ""
description: |-
  Manages CORS (Cross-Origin Resource Sharing) configuration.
---

# agentlink_allowed_origins (Resource)

Manages CORS (Cross-Origin Resource Sharing) configuration for your Frontegg vendor.

## Example Usage

```terraform
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

## Schema

### Required

- `allowed_origins` (Set of String) Set of allowed origin URLs for CORS.

### Read-Only

- `id` (String) The vendor ID.

## Import

Import is supported using the vendor ID:

```shell
terraform import agentlink_allowed_origins.cors <vendor_id>
```
