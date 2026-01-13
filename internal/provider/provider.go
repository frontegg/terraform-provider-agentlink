package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	Region          types.String  `tfsdk:"region"`
	BaseURL         types.String  `tfsdk:"base_url"`
	ClientID        types.String  `tfsdk:"client_id"`
	Secret          types.String  `tfsdk:"secret"`
	ApplicationName types.String  `tfsdk:"application_name"`
	Sources         []SourceModel `tfsdk:"sources"`
}

// SourceModel describes a source configuration
type SourceModel struct {
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	SourceURL  types.String `tfsdk:"source_url"`
	APITimeout types.Int64  `tfsdk:"api_timeout"`
	SchemaFile types.String `tfsdk:"schema_file"`
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
		Description: "Terraform provider for managing Frontegg resources.",
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
			"application_name": schema.StringAttribute{
				Description: "The name of the application as it will appear in the Frontegg Portal. If an application with this name does not exist, it will be created automatically. Can also be set via FRONTEGG_APPLICATION_NAME environment variable.",
				Optional:    true,
			},
			"sources": schema.ListNestedAttribute{
				Description: "List of MCP configuration sources. Each source will be created if it doesn't exist.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the source.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the source. Valid values: REST, GRAPHQL, MOCK, MCP_PROXY, FRONTEGG, CUSTOM_INTEGRATION.",
							Required:    true,
						},
						"source_url": schema.StringAttribute{
							Description: "URL of the source (must be HTTPS).",
							Required:    true,
						},
						"api_timeout": schema.Int64Attribute{
							Description: "API timeout in milliseconds (500-5000). Defaults to 3000.",
							Optional:    true,
						},
						"schema_file": schema.StringAttribute{
							Description: "Path to the schema file to import. For REST sources, this should be an OpenAPI specification (JSON/YAML). For GRAPHQL sources, this should be a GraphQL schema file.",
							Optional:    true,
						},
					},
				},
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
	applicationName := os.Getenv("FRONTEGG_APPLICATION_NAME")

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
	if !config.ApplicationName.IsNull() {
		applicationName = config.ApplicationName.ValueString()
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

	// Find or create application if application_name is provided
	if applicationName != "" {
		app, err := c.FindOrCreateApplication(ctx, applicationName, "https://localhost", "https://localhost")
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Find or Create Application",
				fmt.Sprintf("Failed to find or create application '%s': %s", applicationName, err.Error()),
			)
			return
		}
		_ = app // Application ID is stored in the client
	}

	// Find or create sources if provided
	if len(config.Sources) > 0 {
		if c.ApplicationID == "" {
			resp.Diagnostics.AddError(
				"Application Required for Sources",
				"An application_name must be configured to create sources.",
			)
			return
		}

		for _, srcConfig := range config.Sources {
			apiTimeout := int64(3000) // default
			if !srcConfig.APITimeout.IsNull() {
				apiTimeout = srcConfig.APITimeout.ValueInt64()
			}

			source, err := c.FindOrCreateSource(
				ctx,
				c.ApplicationID,
				srcConfig.Name.ValueString(),
				srcConfig.Type.ValueString(),
				srcConfig.SourceURL.ValueString(),
				int(apiTimeout),
			)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Find or Create Source",
					fmt.Sprintf("Failed to find or create source '%s': %s", srcConfig.Name.ValueString(), err.Error()),
				)
				return
			}

			// Import schema if provided for REST or GRAPHQL sources
			if !srcConfig.SchemaFile.IsNull() && srcConfig.SchemaFile.ValueString() != "" {
				sourceType := srcConfig.Type.ValueString()
				if sourceType != "REST" && sourceType != "GRAPHQL" {
					resp.Diagnostics.AddError(
						"Invalid Source Type for Schema Import",
						fmt.Sprintf("Schema import is only supported for REST and GRAPHQL sources, got '%s'", sourceType),
					)
					return
				}

				schemaFilePath := srcConfig.SchemaFile.ValueString()
				schemaContent, err := os.ReadFile(schemaFilePath)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Read Schema File",
						fmt.Sprintf("Failed to read schema file '%s': %s", schemaFilePath, err.Error()),
					)
					return
				}

				// Extract filename from path for the import
				filename := filepath.Base(schemaFilePath)

				err = c.ImportAndUpsertSchema(ctx, c.ApplicationID, source.ID, sourceType, schemaContent, filename)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to Import Schema",
						fmt.Sprintf("Failed to import schema for source '%s': %s", srcConfig.Name.ValueString(), err.Error()),
					)
					return
				}
			}
		}
	}

	// Make the client available to data sources and resources
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *FronteggProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Resources will be added here as they are implemented
	}
}

func (p *FronteggProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
	}
}
