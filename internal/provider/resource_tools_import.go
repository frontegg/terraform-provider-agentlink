package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/frontegg/terraform-provider-agentlink/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ToolsImportResource{}

func NewToolsImportResource() resource.Resource {
	return &ToolsImportResource{}
}

// ToolsImportResource defines the resource implementation.
type ToolsImportResource struct {
	client *client.Client
}

// ToolsImportResourceModel describes the resource data model.
type ToolsImportResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	SourceID      types.String `tfsdk:"source_id"`
	SchemaFile    types.String `tfsdk:"schema_file"`
	SchemaType    types.String `tfsdk:"schema_type"`
	SchemaHash    types.String `tfsdk:"schema_hash"`
	ToolsCount    types.Int64  `tfsdk:"tools_count"`
}

func (r *ToolsImportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tools_import"
}

func (r *ToolsImportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Imports tools from an OpenAPI or GraphQL schema file.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The import ID (composite of app_id and source_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The application ID to import tools into.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_id": schema.StringAttribute{
				Description: "The source ID to associate tools with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_file": schema.StringAttribute{
				Description: "Path to the OpenAPI (JSON/YAML) or GraphQL schema file.",
				Required:    true,
			},
			"schema_type": schema.StringAttribute{
				Description: "The schema type. Valid values: openapi, graphql.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_hash": schema.StringAttribute{
				Description: "SHA256 hash of the schema file contents (used to detect changes).",
				Computed:    true,
			},
			"tools_count": schema.Int64Attribute{
				Description: "Number of tools imported from the schema.",
				Computed:    true,
			},
		},
	}
}

func (r *ToolsImportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client, got something else.",
		)
		return
	}

	r.client = client
}

func (r *ToolsImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ToolsImportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read schema file
	schemaContent, err := os.ReadFile(data.SchemaFile.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("File Error", "Unable to read schema file: "+err.Error())
		return
	}

	// Calculate hash
	hash := sha256.Sum256(schemaContent)
	hashStr := hex.EncodeToString(hash[:])

	// Determine source type based on schema type
	var sourceType string
	switch data.SchemaType.ValueString() {
	case "openapi":
		sourceType = "REST"
	case "graphql":
		sourceType = "GRAPHQL"
	default:
		resp.Diagnostics.AddError("Invalid Schema Type", "schema_type must be 'openapi' or 'graphql'")
		return
	}

	// Import and upsert schema
	filename := filepath.Base(data.SchemaFile.ValueString())
	err = r.client.ImportAndUpsertSchema(
		ctx,
		data.ApplicationID.ValueString(),
		data.SourceID.ValueString(),
		sourceType,
		schemaContent,
		filename,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to import schema: "+err.Error())
		return
	}

	// Set computed values
	data.ID = types.StringValue(data.ApplicationID.ValueString() + ":" + data.SourceID.ValueString())
	data.SchemaHash = types.StringValue(hashStr)
	// Note: We don't know the exact count without querying, set to -1 to indicate unknown
	data.ToolsCount = types.Int64Value(-1)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolsImportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ToolsImportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if schema file still exists and calculate current hash
	schemaContent, err := os.ReadFile(data.SchemaFile.ValueString())
	if err != nil {
		// File doesn't exist anymore, but that's OK - keep state as is
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	hash := sha256.Sum256(schemaContent)
	hashStr := hex.EncodeToString(hash[:])
	data.SchemaHash = types.StringValue(hashStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolsImportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ToolsImportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read schema file
	schemaContent, err := os.ReadFile(data.SchemaFile.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("File Error", "Unable to read schema file: "+err.Error())
		return
	}

	// Calculate hash
	hash := sha256.Sum256(schemaContent)
	hashStr := hex.EncodeToString(hash[:])

	// Determine source type based on schema type
	var sourceType string
	switch data.SchemaType.ValueString() {
	case "openapi":
		sourceType = "REST"
	case "graphql":
		sourceType = "GRAPHQL"
	default:
		resp.Diagnostics.AddError("Invalid Schema Type", "schema_type must be 'openapi' or 'graphql'")
		return
	}

	// Re-import and upsert schema
	filename := filepath.Base(data.SchemaFile.ValueString())
	err = r.client.ImportAndUpsertSchema(
		ctx,
		data.ApplicationID.ValueString(),
		data.SourceID.ValueString(),
		sourceType,
		schemaContent,
		filename,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to import schema: "+err.Error())
		return
	}

	// Update hash
	data.SchemaHash = types.StringValue(hashStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ToolsImportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ToolsImportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete all tools associated with this source
	err := r.client.DeleteToolsBySource(ctx, data.ApplicationID.ValueString(), data.SourceID.ValueString())
	if err != nil {
		// Log warning but don't fail - tools might already be deleted
		resp.Diagnostics.AddWarning("Cleanup Warning", "Unable to delete tools: "+err.Error())
	}
}
