package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestMcpConfigurationResourceHasExpectedSchema(t *testing.T) {
	r := NewMcpConfigurationResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"application_id", "base_url"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check computed attribute
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("expected 'id' attribute in schema")
	}

	// Check optional attribute
	if _, ok := resp.Schema.Attributes["api_timeout"]; !ok {
		t.Error("expected 'api_timeout' attribute in schema")
	}
}

func TestMcpConfigurationResourceMetadata(t *testing.T) {
	r := NewMcpConfigurationResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_mcp_configuration"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestMcpConfigurationResourceImplementsResource(t *testing.T) {
	r := NewMcpConfigurationResource()

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r.(*McpConfigurationResource)
}
