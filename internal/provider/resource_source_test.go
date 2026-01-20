package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSourceResourceHasExpectedSchema(t *testing.T) {
	r := NewSourceResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"application_id", "name", "type", "source_url"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id", "vendor_id"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected computed attribute '%s' in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"api_timeout", "enabled"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected optional attribute '%s' in schema", attr)
		}
	}
}

func TestSourceResourceMetadata(t *testing.T) {
	r := NewSourceResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_source"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestSourceResourceImplementsResource(t *testing.T) {
	r := NewSourceResource()

	var _ = r
	var _ resource.ResourceWithImportState = r.(*SourceResource)
}
