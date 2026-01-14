package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestToolsImportResourceHasExpectedSchema(t *testing.T) {
	r := NewToolsImportResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"application_id", "source_id", "schema_file", "schema_type"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id", "schema_hash", "tools_count"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected computed attribute '%s' in schema", attr)
		}
	}
}

func TestToolsImportResourceMetadata(t *testing.T) {
	r := NewToolsImportResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_tools_import"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestToolsImportResourceImplementsResource(t *testing.T) {
	r := NewToolsImportResource()

	var _ resource.Resource = r
	// Note: ToolsImportResource does not implement ImportState
}
