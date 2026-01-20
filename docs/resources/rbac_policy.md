---
page_title: "agentlink_rbac_policy Resource - AgentLink"
subcategory: ""
description: |-
  Manages Role-Based Access Control (RBAC) policies for tools.
---

# agentlink_rbac_policy (Resource)

Manages Role-Based Access Control (RBAC) policies. Restrict which tools can be accessed based on user roles or permissions.

## Example Usage

### Restrict by Roles

```terraform
resource "agentlink_rbac_policy" "admin_tools" {
  name              = "Admin Only Tools"
  description       = "Restrict sensitive tools to admin users"
  enabled           = true
  type              = "RBAC_ROLES"
  keys              = ["admin", "super-admin"]
  internal_tool_ids = ["tool-id-1", "tool-id-2"]
  app_ids           = [agentlink_application.main.id]
}
```

### Restrict by Permissions

```terraform
resource "agentlink_rbac_policy" "write_permission" {
  name              = "Write Permission Required"
  description       = "Tools that modify data require write permission"
  enabled           = true
  type              = "RBAC_PERMISSIONS"
  keys              = ["data:write", "data:admin"]
  internal_tool_ids = []  # Empty = apply to all tools
}
```

## Schema

### Required

- `name` (String) Policy name.
- `enabled` (Boolean) Whether the policy is enabled.
- `type` (String) RBAC type. Valid values: `RBAC_ROLES`, `RBAC_PERMISSIONS`. Changing this forces a new resource to be created.
- `keys` (List of String) List of role or permission keys. At least one required.
- `internal_tool_ids` (List of String) List of tool IDs. Empty list applies to all tools.

### Optional

- `description` (String) Policy description.
- `app_ids` (List of String) List of application IDs to apply policy to.
- `tenant_id` (String) Tenant ID for multi-tenant scenarios.

### Read-Only

- `id` (String) The policy ID.

## Import

Import is supported using the policy ID:

```shell
terraform import agentlink_rbac_policy.admin_tools <policy_id>
```
