package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestAllowedOriginsResourceHasExpectedSchema(t *testing.T) {
	r := NewAllowedOriginsResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.TODO(), resource.SchemaRequest{}, schemaResp)

	if schemaResp.Schema.Attributes == nil {
		t.Fatal("Schema should have attributes")
	}

	// Check required attributes (allowed_origins is a Set)
	requiredAttrs := []string{"allowed_origins"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema should have attribute: %s", attr)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id"}
	for _, attr := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema should have computed attribute: %s", attr)
		}
	}

	// Verify allowed_origins is a SetAttribute
	if _, ok := schemaResp.Schema.Attributes["allowed_origins"].(schema.SetAttribute); !ok {
		t.Error("allowed_origins should be a SetAttribute")
	}
}

func TestAllowedOriginsResourceMetadata(t *testing.T) {
	r := NewAllowedOriginsResource()

	metaResp := &resource.MetadataResponse{}
	r.Metadata(context.TODO(), resource.MetadataRequest{ProviderTypeName: "agentlink"}, metaResp)

	if metaResp.TypeName != "agentlink_allowed_origins" {
		t.Errorf("Expected type name agentlink_allowed_origins, got %s", metaResp.TypeName)
	}
}

func TestAllowedOriginsResourceImplementsResource(t *testing.T) {
	var _ = NewAllowedOriginsResource()
	var _ resource.ResourceWithImportState = NewAllowedOriginsResource().(*AllowedOriginsResource)
}
