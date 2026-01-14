package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-frontegg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &McpConfigurationResource{}
var _ resource.ResourceWithImportState = &McpConfigurationResource{}

func NewMcpConfigurationResource() resource.Resource {
	return &McpConfigurationResource{}
}

// McpConfigurationResource defines the resource implementation.
type McpConfigurationResource struct {
	client *client.Client
}

// McpConfigurationResourceModel describes the resource data model.
type McpConfigurationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	BaseURL       types.String `tfsdk:"base_url"`
	APITimeout    types.Int64  `tfsdk:"api_timeout"`
}

func (r *McpConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_configuration"
}

func (r *McpConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages MCP (Model Context Protocol) configuration for an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The MCP configuration ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The application ID this configuration belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"base_url": schema.StringAttribute{
				Description: "The base URL for the API that tools will call.",
				Required:    true,
			},
			"api_timeout": schema.Int64Attribute{
				Description: "API timeout in milliseconds. Defaults to 5000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(5000),
			},
		},
	}
}

func (r *McpConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *McpConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data McpConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateOrUpdateMcpConfigurationRequest{
		AppID:      data.ApplicationID.ValueString(),
		BaseURL:    data.BaseURL.ValueString(),
		APITimeout: int(data.APITimeout.ValueInt64()),
	}

	config, err := r.client.CreateOrUpdateMcpConfiguration(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create MCP configuration: "+err.Error())
		return
	}

	data.ID = types.StringValue(config.ID)
	data.ApplicationID = types.StringValue(config.AppID)
	data.BaseURL = types.StringValue(config.BaseURL)
	data.APITimeout = types.Int64Value(int64(config.APITimeout))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *McpConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data McpConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetMcpConfiguration(ctx, data.ApplicationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read MCP configuration: "+err.Error())
		return
	}

	if config == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(config.ID)
	data.ApplicationID = types.StringValue(config.AppID)
	data.BaseURL = types.StringValue(config.BaseURL)
	data.APITimeout = types.Int64Value(int64(config.APITimeout))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *McpConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data McpConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.CreateOrUpdateMcpConfigurationRequest{
		AppID:      data.ApplicationID.ValueString(),
		BaseURL:    data.BaseURL.ValueString(),
		APITimeout: int(data.APITimeout.ValueInt64()),
	}

	config, err := r.client.CreateOrUpdateMcpConfiguration(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update MCP configuration: "+err.Error())
		return
	}

	data.ID = types.StringValue(config.ID)
	data.ApplicationID = types.StringValue(config.AppID)
	data.BaseURL = types.StringValue(config.BaseURL)
	data.APITimeout = types.Int64Value(int64(config.APITimeout))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *McpConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// MCP Configuration doesn't have a delete endpoint - it's tied to the application
	// Just remove from state
}

func (r *McpConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by application_id
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}
