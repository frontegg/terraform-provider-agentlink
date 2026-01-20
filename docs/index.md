---
page_title: "Provider: AgentLink"
description: |-
  The AgentLink provider is used to manage Frontegg AgentLink resources for building secure AI agents.
---

# AgentLink Provider

The AgentLink provider is used to manage [Frontegg AgentLink](https://frontegg.com) resources. AgentLink enables you to build and deploy secure AI agents with enterprise-grade access control, data masking, and Model Context Protocol (MCP) integration.

## What is Frontegg AgentLink?

Frontegg AgentLink is a platform for building secure AI agents that can interact with your APIs and services. It provides:

- **Model Context Protocol (MCP) Support**: Standard protocol for AI model interactions with tools and APIs
- **Enterprise Security**: Role-based access control (RBAC), data masking, and conditional policies
- **Multi-Source Integration**: Connect REST APIs, GraphQL endpoints, and custom integrations
- **Tool Management**: Import tools from OpenAPI/GraphQL schemas with automatic discovery

## Example Usage

```terraform
terraform {
  required_providers {
    agentlink = {
      source  = "frontegg/agentlink"
      version = "~> 0.1"
    }
  }
}

provider "agentlink" {
  region = "eu"
}

# Create an AI agent application
resource "agentlink_application" "my_agent" {
  name      = "My AI Agent"
  app_url   = "https://agent.example.com"
  login_url = "https://agent.example.com/login"
  type      = "agent"
  allow_dcr = true
}

# Configure MCP settings
resource "agentlink_mcp_configuration" "config" {
  application_id = agentlink_application.my_agent.id
  base_url       = "https://api.example.com"
}
```

## Authentication

The provider requires Frontegg API credentials. Obtain these from the [Frontegg Portal](https://portal.frontegg.com) under **Settings > API Keys**.

### Environment Variables (Recommended)

```bash
export FRONTEGG_CLIENT_ID="your-client-id"
export FRONTEGG_SECRET="your-secret"
export FRONTEGG_REGION="eu"  # Optional, defaults to "eu"
```

### Provider Configuration

```terraform
provider "agentlink" {
  client_id = "your-client-id"
  secret    = "your-secret"
  region    = "eu"
}
```

## Schema

### Optional

- `client_id` (String) Frontegg API client ID. Can also be set via `FRONTEGG_CLIENT_ID` environment variable.
- `secret` (String, Sensitive) Frontegg API secret. Can also be set via `FRONTEGG_SECRET` environment variable.
- `region` (String) Frontegg region. Defaults to `eu`. Can also be set via `FRONTEGG_REGION` environment variable.
- `base_url` (String) Override API base URL. Normally derived from region.

### Supported Regions

| Region | API Endpoint |
|--------|--------------|
| `stg` | https://api.stg.frontegg.com |
| `eu` | https://api.frontegg.com |
| `us` | https://api.us.frontegg.com |
| `au` | https://api.au.frontegg.com |
| `ca` | https://api.ca.frontegg.com |
| `uk` | https://api.uk.frontegg.com |
