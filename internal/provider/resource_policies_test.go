package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ============================================================================
// Conditional Policy Tests
// ============================================================================

func TestConditionalPolicyResourceHasExpectedSchema(t *testing.T) {
	r := NewConditionalPolicyResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"name", "enabled", "internal_tool_ids"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"description", "app_ids", "tenant_id", "targeting", "metadata"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected optional attribute '%s' in schema", attr)
		}
	}

	// Check computed attribute
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("expected 'id' attribute in schema")
	}
}

func TestConditionalPolicyResourceMetadata(t *testing.T) {
	r := NewConditionalPolicyResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_conditional_policy"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestConditionalPolicyResourceImplementsResource(t *testing.T) {
	r := NewConditionalPolicyResource()

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r.(*ConditionalPolicyResource)
}

// ============================================================================
// RBAC Policy Tests
// ============================================================================

func TestRbacPolicyResourceHasExpectedSchema(t *testing.T) {
	r := NewRbacPolicyResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"name", "enabled", "type", "keys", "internal_tool_ids"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"description", "app_ids", "tenant_id"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected optional attribute '%s' in schema", attr)
		}
	}
}

func TestRbacPolicyResourceMetadata(t *testing.T) {
	r := NewRbacPolicyResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_rbac_policy"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestRbacPolicyResourceImplementsResource(t *testing.T) {
	r := NewRbacPolicyResource()

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r.(*RbacPolicyResource)
}

// ============================================================================
// Masking Policy Tests
// ============================================================================

func TestMaskingPolicyResourceHasExpectedSchema(t *testing.T) {
	r := NewMaskingPolicyResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check required attributes
	requiredAttrs := []string{"name", "enabled", "internal_tool_ids", "policy_configuration"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}

	// Check optional attributes
	optionalAttrs := []string{"description", "app_ids", "tenant_id"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected optional attribute '%s' in schema", attr)
		}
	}
}

func TestMaskingPolicyResourceMetadata(t *testing.T) {
	r := NewMaskingPolicyResource()

	req := resource.MetadataRequest{ProviderTypeName: "agentlink"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "agentlink_masking_policy"
	if resp.TypeName != expected {
		t.Errorf("expected type name '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestMaskingPolicyResourceImplementsResource(t *testing.T) {
	r := NewMaskingPolicyResource()

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r.(*MaskingPolicyResource)
}
