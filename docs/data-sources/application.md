---
page_title: "agentlink_application Data Source - AgentLink"
subcategory: ""
description: |-
  Retrieves information about the current application.
---

# agentlink_application (Data Source)

Retrieves information about the current application configured for the provider.

## Example Usage

```terraform
data "agentlink_application" "current" {}

output "application_info" {
  value = {
    id   = data.agentlink_application.current.id
    name = data.agentlink_application.current.name
  }
}
```

## Schema

### Read-Only

- `id` (String) The application ID.
- `name` (String) The application name.
- `app_url` (String) The application URL.
- `login_url` (String) The login/OAuth URL.
- `type` (String) The application type.
- `access_type` (String) The access type.
- `allow_dcr` (Boolean) Whether Dynamic Client Registration is enabled.
- `description` (String) The application description.
- `is_active` (Boolean) Whether the application is active.
- `is_default` (Boolean) Whether this is the default application.
- `logo_url` (String) The application logo URL.
- `frontend_stack` (String) The frontend framework.
- `vendor_id` (String) The vendor ID.
- `app_host` (String) The application host.
