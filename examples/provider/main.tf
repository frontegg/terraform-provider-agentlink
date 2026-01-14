terraform {
  required_providers {
    agentlink = {
      source = "frontegg/agentlink"
    }
  }
}

variable "client_id" {
  description = "Frontegg Client ID"
  type        = string
  sensitive   = true
}

variable "secret" {
  description = "Frontegg Secret"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "Frontegg region: stg, eu (default), us, au, ca, uk"
  type        = string
  default     = "eu"
}

# Provider configuration - authentication only
provider "agentlink" {
  client_id = var.client_id
  secret    = var.secret
  region    = var.region
}

# ============================================================================
# Application Resource
# ============================================================================

resource "agentlink_application" "main" {
  name        = "My MCP Server"
  app_url     = "https://app.example.com"
  login_url   = "https://app.example.com/oauth"
  type        = "agent"
  allow_dcr   = true
  description = "My MCP server application"
}

# ============================================================================
# MCP Configuration
# ============================================================================

resource "agentlink_mcp_configuration" "main" {
  application_id = agentlink_application.main.id
  base_url       = "https://api.example.com"
  api_timeout    = 5000
}

# ============================================================================
# Source Configuration
# ============================================================================

resource "agentlink_source" "rest_api" {
  application_id = agentlink_application.main.id
  name           = "My REST API"
  type           = "REST"
  source_url     = "https://api.example.com"
  api_timeout    = 3000
  enabled        = true
}

# ============================================================================
# Tools Import (OpenAPI)
# ============================================================================

resource "agentlink_tools_import" "openapi" {
  application_id = agentlink_application.main.id
  source_id      = agentlink_source.rest_api.id
  schema_file    = "./sample-openapi.json"
  schema_type    = "openapi"
}

# ============================================================================
# RBAC Policy - Role-based access control
# ============================================================================

# resource "agentlink_rbac_policy" "admin_only" {
#   name              = "Admin Only Tools"
#   description       = "Restrict sensitive tools to admin users only"
#   enabled           = true
#   type              = "RBAC_ROLES"
#   keys              = ["admin", "super-admin"]
#   internal_tool_ids = []  # Would reference actual tool IDs
# }

# ============================================================================
# Masking Policy - Data protection
# ============================================================================

# resource "agentlink_masking_policy" "pii_protection" {
#   name              = "PII Data Masking"
#   description       = "Mask personally identifiable information in responses"
#   enabled           = true
#   internal_tool_ids = []  # Apply to all tools
#
#   policy_configuration {
#     credit_card   = true
#     email_address = true
#     phone_number  = true
#     us_ssn        = true
#     ip_address    = false
#   }
# }

# ============================================================================
# Outputs
# ============================================================================

output "application_id" {
  description = "The created application ID"
  value       = agentlink_application.main.id
}

output "application_name" {
  description = "The application name"
  value       = agentlink_application.main.name
}

output "mcp_configuration_id" {
  description = "The MCP configuration ID"
  value       = agentlink_mcp_configuration.main.id
}

output "source_id" {
  description = "The source ID"
  value       = agentlink_source.rest_api.id
}
