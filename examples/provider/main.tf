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

variable "application_name" {
  description = "Name of the application as it will appear in the Frontegg Portal. If not found, it will be created automatically."
  type        = string
}

provider "agentlink" {
  client_id        = var.client_id
  secret           = var.secret
  region           = var.region
  application_name = var.application_name

  sources = [
    {
      name        = "My REST API"
      type        = "REST"
      source_url  = "https://example.com"
      api_timeout = 3000
      schema_file = "./sample-openapi.json"
    }
  ]
}

# Data source to trigger provider configuration and get application info
data "agentlink_application" "current" {}

output "application_id" {
  value = data.agentlink_application.current.id
}
