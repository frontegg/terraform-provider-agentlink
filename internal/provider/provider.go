package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/frontegg/terraform-provider-frontegg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure FronteggProvider satisfies various provider interfaces.
var _ provider.Provider = &FronteggProvider{}

// regionURLs maps region identifiers to their API base URLs
var regionURLs = map[string]string{
	"stg": "https://api.stg.frontegg.com",
	"eu":  "https://api.frontegg.com",
	"us":  "https://api.us.frontegg.com",
	"au":  "https://api.au.frontegg.com",
	"ca":  "https://api.ca.frontegg.com",
	"uk":  "https://api.uk.frontegg.com",
}

// validRegions returns a list of valid region names for error messages
func validRegions() []string {
	return []string{"stg", "eu", "us", "au", "ca", "uk"}
}

// FronteggProvider defines the provider implementation.
type FronteggProvider struct {
	version string
}

// FronteggProviderModel describes the provider data model.
type FronteggProviderModel struct {
	Region   types.String `tfsdk:"region"`
	BaseURL  types.String `tfsdk:"base_url"`
	ClientID types.String `tfsdk:"client_id"`
	Secret   types.String `tfsdk:"secret"`
}

// New creates a new provider factory function
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FronteggProvider{
			version: version,
		}
	}
}

func (p *FronteggProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "agentlink"
	resp.Version = p.version
}

func (p *FronteggProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Frontegg AgentLink resources including applications, MCP configurations, sources, tools, and policies.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "The Frontegg region. Valid values: stg, eu (default), us, au, ca, uk. Can also be set via FRONTEGG_REGION environment variable.",
				Optional:    true,
			},
			"base_url": schema.StringAttribute{
				Description: "Override the base URL for the Frontegg API. If set, this takes precedence over region. Can also be set via FRONTEGG_BASE_URL environment variable.",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "The client ID for Frontegg API authentication. Can also be set via FRONTEGG_CLIENT_ID environment variable.",
				Optional:    true,
			},
			"secret": schema.StringAttribute{
				Description: "The secret for Frontegg API authentication. Can also be set via FRONTEGG_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *FronteggProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FronteggProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for unknown values
	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Frontegg Client ID",
			"The provider cannot create the Frontegg API client as there is an unknown configuration value for the Frontegg client ID.",
		)
	}

	if config.Secret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Unknown Frontegg Secret",
			"The provider cannot create the Frontegg API client as there is an unknown configuration value for the Frontegg secret.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables
	region := os.Getenv("FRONTEGG_REGION")
	baseURL := os.Getenv("FRONTEGG_BASE_URL")
	clientID := os.Getenv("FRONTEGG_CLIENT_ID")
	secret := os.Getenv("FRONTEGG_SECRET")

	// Override with config values if provided
	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}
	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueString()
	}
	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}

	// Resolve base URL: base_url takes precedence over region
	if baseURL == "" {
		if region == "" {
			region = "eu" // Default region
		}
		url, ok := regionURLs[region]
		if !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("region"),
				"Invalid Frontegg Region",
				fmt.Sprintf("The region '%s' is not valid. Valid regions are: %v", region, validRegions()),
			)
			return
		}
		baseURL = url
	}

	// Validate required values
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Frontegg Client ID",
			"The provider requires a client_id to authenticate with the Frontegg API. "+
				"Set the client_id value in the provider configuration or use the FRONTEGG_CLIENT_ID environment variable.",
		)
	}

	if secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Missing Frontegg Secret",
			"The provider requires a secret to authenticate with the Frontegg API. "+
				"Set the secret value in the provider configuration or use the FRONTEGG_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create client
	c := client.NewClient(baseURL, clientID, secret)

	// Verify authentication
	if err := c.Authenticate(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate with Frontegg API",
			"An unexpected error occurred when authenticating with the Frontegg API. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to data sources and resources
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *FronteggProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
		NewMcpConfigurationResource,
		NewSourceResource,
		NewToolsImportResource,
		NewConditionalPolicyResource,
		NewRbacPolicyResource,
		NewMaskingPolicyResource,
	}
}

func (p *FronteggProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
	}
}
