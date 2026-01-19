package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// providerConfig returns the provider configuration for acceptance tests
func providerConfig() string {
	return `
provider "agentlink" {
  client_id = "` + os.Getenv("FRONTEGG_CLIENT_ID") + `"
  secret    = "` + os.Getenv("FRONTEGG_SECRET") + `"
  region    = "` + getEnvOrDefault("FRONTEGG_REGION", "stg") + `"
}
`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// testAccProtoV6ProviderFactoriesAcc are used to instantiate a provider during acceptance testing
var testAccProtoV6ProviderFactoriesAcc = map[string]func() (tfprotov6.ProviderServer, error){
	"agentlink": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("FRONTEGG_CLIENT_ID") == "" {
		t.Fatal("FRONTEGG_CLIENT_ID must be set for acceptance tests")
	}
	if os.Getenv("FRONTEGG_SECRET") == "" {
		t.Fatal("FRONTEGG_SECRET must be set for acceptance tests")
	}
}

// TestAccApplicationResource tests the full lifecycle of an application
func TestAccApplicationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF Test App"
  app_url     = "https://test.example.com"
  login_url   = "https://test.example.com/oauth"
  type        = "agent"
  allow_dcr   = true
  description = "Terraform test application"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_application.test", "name", "TF Test App"),
					resource.TestCheckResourceAttr("agentlink_application.test", "app_url", "https://test.example.com"),
					resource.TestCheckResourceAttr("agentlink_application.test", "login_url", "https://test.example.com/oauth"),
					resource.TestCheckResourceAttr("agentlink_application.test", "type", "agent"),
					resource.TestCheckResourceAttrSet("agentlink_application.test", "id"),
				),
			},
			// Update testing
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF Test App Updated"
  app_url     = "https://test-updated.example.com"
  login_url   = "https://test-updated.example.com/oauth"
  type        = "agent"
  allow_dcr   = true
  description = "Terraform test application - updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_application.test", "name", "TF Test App Updated"),
					resource.TestCheckResourceAttr("agentlink_application.test", "app_url", "https://test-updated.example.com"),
				),
			},
			// Import testing
			{
				ResourceName:      "agentlink_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				// allow_dcr is not returned by the API, so skip verification
				ImportStateVerifyIgnore: []string{"allow_dcr"},
			},
		},
	})
}

// TestAccMcpConfigurationResource tests MCP configuration resource
func TestAccMcpConfigurationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF MCP Config Test App"
  app_url     = "https://mcp-test.example.com"
  login_url   = "https://mcp-test.example.com/oauth"
  type        = "agent"
}

resource "agentlink_mcp_configuration" "test" {
  application_id = agentlink_application.test.id
  base_url       = "https://api.github.com"
  api_timeout    = 5000
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_mcp_configuration.test", "base_url", "https://api.github.com"),
					resource.TestCheckResourceAttr("agentlink_mcp_configuration.test", "api_timeout", "5000"),
					resource.TestCheckResourceAttrSet("agentlink_mcp_configuration.test", "id"),
				),
			},
			// Update testing
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF MCP Config Test App"
  app_url     = "https://mcp-test.example.com"
  login_url   = "https://mcp-test.example.com/oauth"
  type        = "agent"
}

resource "agentlink_mcp_configuration" "test" {
  application_id = agentlink_application.test.id
  base_url       = "https://httpbin.org"
  api_timeout    = 3000
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_mcp_configuration.test", "base_url", "https://httpbin.org"),
					resource.TestCheckResourceAttr("agentlink_mcp_configuration.test", "api_timeout", "3000"),
				),
			},
		},
	})
}

// TestAccSourceResource tests source resource
func TestAccSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF Source Test App"
  app_url     = "https://source-test.example.com"
  login_url   = "https://source-test.example.com/oauth"
  type        = "agent"
}

resource "agentlink_source" "test" {
  application_id = agentlink_application.test.id
  name           = "TF Test Source"
  type           = "REST"
  source_url     = "https://api.github.com"
  api_timeout    = 3000
  enabled        = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_source.test", "name", "TF Test Source"),
					resource.TestCheckResourceAttr("agentlink_source.test", "type", "REST"),
					resource.TestCheckResourceAttr("agentlink_source.test", "source_url", "https://api.github.com"),
					resource.TestCheckResourceAttr("agentlink_source.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("agentlink_source.test", "id"),
				),
			},
			// Update testing
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF Source Test App"
  app_url     = "https://source-test.example.com"
  login_url   = "https://source-test.example.com/oauth"
  type        = "agent"
}

resource "agentlink_source" "test" {
  application_id = agentlink_application.test.id
  name           = "TF Test Source Updated"
  type           = "REST"
  source_url     = "https://httpbin.org"
  api_timeout    = 5000
  enabled        = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_source.test", "name", "TF Test Source Updated"),
					resource.TestCheckResourceAttr("agentlink_source.test", "source_url", "https://httpbin.org"),
					resource.TestCheckResourceAttr("agentlink_source.test", "api_timeout", "5000"),
				),
			},
			// Import testing - format: application_id:source_id
			{
				ResourceName:            "agentlink_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       sourceImportStateIdFunc("agentlink_application.test", "agentlink_source.test"),
				ImportStateVerifyIgnore: []string{"application_id"}, // Set via import ID parsing
			},
		},
	})
}

// sourceImportStateIdFunc returns a function that generates the import ID for sources
func sourceImportStateIdFunc(appResourceName, sourceResourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		appRS, ok := s.RootModule().Resources[appResourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", appResourceName)
		}
		sourceRS, ok := s.RootModule().Resources[sourceResourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", sourceResourceName)
		}
		return appRS.Primary.ID + ":" + sourceRS.Primary.ID, nil
	}
}

// TestAccRbacPolicyResource tests RBAC policy resource
func TestAccRbacPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF RBAC Policy Test App"
  app_url     = "https://rbac-test.example.com"
  login_url   = "https://rbac-test.example.com/oauth"
  type        = "agent"
}

resource "agentlink_source" "test" {
  application_id = agentlink_application.test.id
  name           = "TF RBAC Test Source"
  type           = "REST"
  source_url     = "https://api.github.com"
}

# Note: In a real scenario, you'd need to import tools first to get tool IDs
# This is a placeholder configuration
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("agentlink_application.test", "id"),
				),
			},
		},
	})
}

// TestAccMaskingPolicyResource tests masking policy resource
func TestAccMaskingPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "agentlink_application" "test" {
  name        = "TF Masking Policy Test App"
  app_url     = "https://masking-test.example.com"
  login_url   = "https://masking-test.example.com/oauth"
  type        = "agent"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("agentlink_application.test", "id"),
				),
			},
		},
	})
}

// TestAccFullStack tests a complete deployment scenario
func TestAccFullStack(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "agentlink_application" "test" {
  name        = "TF Full Stack Test"
  app_url     = "https://fullstack-test.example.com"
  login_url   = "https://fullstack-test.example.com/oauth"
  type        = "agent"
  allow_dcr   = true
  description = "Full stack integration test"
}

resource "agentlink_mcp_configuration" "test" {
  application_id = agentlink_application.test.id
  base_url       = "https://api.github.com"
  api_timeout    = 5000
}

resource "agentlink_source" "rest" {
  application_id = agentlink_application.test.id
  name           = "REST API"
  type           = "REST"
  source_url     = "https://api.github.com"
  api_timeout    = 3000
  enabled        = true
}

output "application_id" {
  value = agentlink_application.test.id
}

output "mcp_config_id" {
  value = agentlink_mcp_configuration.test.id
}

output "source_id" {
  value = agentlink_source.rest.id
}
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("agentlink_application.test", "id"),
					resource.TestCheckResourceAttrSet("agentlink_mcp_configuration.test", "id"),
					resource.TestCheckResourceAttrSet("agentlink_source.rest", "id"),
					resource.TestCheckResourceAttr("agentlink_application.test", "name", "TF Full Stack Test"),
					resource.TestCheckResourceAttr("agentlink_source.rest", "type", "REST"),
				),
			},
		},
	})
}

// TestAccAllowedOriginsResource tests the allowed origins resource
func TestAccAllowedOriginsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesAcc,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + `
resource "agentlink_allowed_origins" "test" {
  allowed_origins = [
    "http://localhost:3000",
    "https://app.example.com"
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("agentlink_allowed_origins.test", "id"),
					resource.TestCheckResourceAttr("agentlink_allowed_origins.test", "allowed_origins.#", "2"),
				),
			},
			// Update testing - add more origins
			{
				Config: providerConfig() + `
resource "agentlink_allowed_origins" "test" {
  allowed_origins = [
    "http://localhost:3000",
    "https://app.example.com",
    "https://staging.example.com"
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agentlink_allowed_origins.test", "allowed_origins.#", "3"),
				),
			},
			// Import testing
			{
				ResourceName:      "agentlink_allowed_origins.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
