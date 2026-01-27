package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderHasExpectedResources(t *testing.T) {
	p := &FronteggProvider{}
	resources := p.Resources(context.Background())

	expectedCount := 9
	if len(resources) != expectedCount {
		t.Errorf("expected %d resources, got %d", expectedCount, len(resources))
	}
}

func TestProviderHasExpectedDataSources(t *testing.T) {
	p := &FronteggProvider{}
	dataSources := p.DataSources(context.Background())

	if len(dataSources) < 1 {
		t.Error("expected at least 1 data source")
	}
}

func TestNewProviderFactory(t *testing.T) {
	factory := New("1.0.0")
	p := factory()

	if p == nil {
		t.Fatal("expected provider, got nil")
	}

	// Verify the provider is non-nil (the factory returns provider.Provider)
	_ = p // Type is already provider.Provider from factory return type
}

func TestProviderMetadataTypeName(t *testing.T) {
	p := New("1.0.0")()

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "agentlink" {
		t.Errorf("expected TypeName 'agentlink', got '%s'", resp.TypeName)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got '%s'", resp.Version)
	}
}

func TestProviderSchemaHasRequiredAttributes(t *testing.T) {
	p := New("test")()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check required attributes exist
	requiredAttrs := []string{"client_id", "secret", "region", "base_url"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute '%s' in schema", attr)
		}
	}
}
