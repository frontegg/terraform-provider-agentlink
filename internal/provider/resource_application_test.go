package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestApplicationResourceHasExpectedSchema(t *testing.T) {
	r := NewApplicationResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"name", "app_url", "login_url"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id", "vendor_id", "app_host"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected computed attribute '%s' in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"type", "access_type", "is_active", "allow_dcr", "description", "frontend_stack"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected optional attribute '%s' in schema", attr)
		}
	}
}

func TestApplicationResourceMetadata(t *testing.T) {
	r := NewApplicationResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_application"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestApplicationResourceImplementsResource(t *testing.T) {
	r := NewApplicationResource()

	// Verify it implements the resource.Resource interface
	var _ = r

	// Verify it implements ResourceWithImportState
	var _ resource.ResourceWithImportState = r.(*ApplicationResource)
}
