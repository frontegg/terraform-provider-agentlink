---
page_title: "agentlink_masking_policy Resource - AgentLink"
subcategory: ""
description: |-
  Manages data masking policies for protecting sensitive information.
---

# agentlink_masking_policy (Resource)

Manages data masking policies for protecting sensitive information like PII, financial data, and crypto addresses. Masking policies automatically redact sensitive data in tool responses.

## Example Usage

```terraform
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

## Schema

### Required

- `name` (String) Policy name.
- `enabled` (Boolean) Whether the policy is enabled.
- `internal_tool_ids` (List of String) List of tool IDs. Empty list applies to all tools.
- `policy_configuration` (Block) Masking configuration. See below.

### Optional

- `description` (String) Policy description.
- `app_ids` (List of String) List of application IDs.
- `tenant_id` (String) Tenant ID.

### Read-Only

- `id` (String) The policy ID.

### Nested Schema for `policy_configuration`

All fields are optional booleans that default to `false`:

- `credit_card` - Mask credit card numbers
- `email_address` - Mask email addresses
- `phone_number` - Mask phone numbers
- `ip_address` - Mask IP addresses
- `us_ssn` - Mask US Social Security Numbers
- `us_driver_license` - Mask US driver licenses
- `us_passport` - Mask US passport numbers
- `us_itin` - Mask US Individual Taxpayer Identification Numbers
- `us_bank_number` - Mask US bank account numbers
- `iban_code` - Mask International Bank Account Numbers
- `swift_code` - Mask SWIFT/BIC codes
- `bitcoin_address` - Mask Bitcoin wallet addresses
- `ethereum_address` - Mask Ethereum wallet addresses
- `cvv_cvc` - Mask credit card CVV/CVC codes
- `url` - Mask URLs

## Import

Import is supported using the policy ID:

```shell
terraform import agentlink_masking_policy.pii_protection <policy_id>
```
