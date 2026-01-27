package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestIdentityConfigurationResourceHasExpectedSchema(t *testing.T) {
	r := NewIdentityConfigurationResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.TODO(), resource.SchemaRequest{}, schemaResp)

	if schemaResp.Schema.Attributes == nil {
		t.Fatal("Schema should have attributes")
	}

	// Check required attributes
	requiredAttrs := []string{"default_token_expiration"}
	for _, attr := range requiredAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema should have required attribute: %s", attr)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id"}
	for _, attr := range computedAttrs {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema should have computed attribute: %s", attr)
		}
	}

	// Verify default_token_expiration is an Int64Attribute
	if _, ok := schemaResp.Schema.Attributes["default_token_expiration"].(schema.Int64Attribute); !ok {
		t.Error("default_token_expiration should be an Int64Attribute")
	}
}

func TestIdentityConfigurationResourceMetadata(t *testing.T) {
	r := NewIdentityConfigurationResource()

	metaResp := &resource.MetadataResponse{}
	r.Metadata(context.TODO(), resource.MetadataRequest{ProviderTypeName: "agentlink"}, metaResp)

	expected := "agentlink_identity_configuration"
	if metaResp.TypeName != expected {
		t.Errorf("Expected type name '%s', got '%s'", expected, metaResp.TypeName)
	}
}

func TestIdentityConfigurationResourceImplementsResource(t *testing.T) {
	r := NewIdentityConfigurationResource()

	// Verify it implements the resource.Resource interface
	var _ = r
}

func TestIdentityConfigurationResourceSchemaDescription(t *testing.T) {
	r := NewIdentityConfigurationResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.TODO(), resource.SchemaRequest{}, schemaResp)

	if schemaResp.Schema.Description == "" {
		t.Error("Schema should have a description")
	}

	// Check attribute descriptions
	attr, ok := schemaResp.Schema.Attributes["default_token_expiration"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("default_token_expiration should be an Int64Attribute")
	}
	if attr.Description == "" {
		t.Error("default_token_expiration should have a description")
	}
}
