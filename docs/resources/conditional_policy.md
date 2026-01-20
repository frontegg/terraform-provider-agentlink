---
page_title: "agentlink_conditional_policy Resource - AgentLink"
subcategory: ""
description: |-
  Manages conditional policies with advanced targeting rules.
---

# agentlink_conditional_policy (Resource)

Manages conditional policies with advanced targeting rules. Use these for complex access control scenarios, approval workflows, and context-aware security.

## Example Usage

### Require Approval for Destructive Operations

```terraform
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
```

### Deny Access Outside Business Hours

```terraform
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
```

### Allow Specific Users

```terraform
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

## Schema

### Required

- `name` (String) Policy name.
- `enabled` (Boolean) Whether the policy is enabled.
- `internal_tool_ids` (List of String) List of tool IDs. Empty list applies to all tools.

### Optional

- `description` (String) Policy description.
- `targeting` (Block) Targeting rules. See below.
- `app_ids` (List of String) List of application IDs.
- `tenant_id` (String) Tenant ID.
- `metadata` (Map of String) Additional metadata.

### Read-Only

- `id` (String) The policy ID.

### Nested Schema for `targeting`

#### `if` Block

- `condition` (Block) One or more condition blocks.

#### `condition` Block

- `attribute` (String) The attribute to evaluate (e.g., `tool.method`, `user.email`, `request.hour`).
- `negate` (Boolean) Whether to negate the condition.
- `op` (String) Operator. Valid values: `equals`, `not_equals`, `in_list`, `not_in_list`, `in_range`, `not_in_range`.
- `value` (Map) Value map with keys like `string`, `list`, `range_start`, `range_end`.

#### `then` Block

- `result` (String) Result action. Valid values: `ALLOW`, `DENY`, `APPROVAL_REQUIRED`.
- `approval_flow_id` (String) Approval flow ID (required when result is `APPROVAL_REQUIRED`).

## Import

Import is supported using the policy ID:

```shell
terraform import agentlink_conditional_policy.delete_approval <policy_id>
```
