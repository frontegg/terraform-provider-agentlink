package provider

import (
	"context"
	"strings"

	"github.com/frontegg/terraform-provider-agentlink/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SourceResource{}
var _ resource.ResourceWithImportState = &SourceResource{}

func NewSourceResource() resource.Resource {
	return &SourceResource{}
}

// SourceResource defines the resource implementation.
type SourceResource struct {
	client *client.Client
}

// SourceResourceModel describes the resource data model.
type SourceResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	SourceURL     types.String `tfsdk:"source_url"`
	APITimeout    types.Int64  `tfsdk:"api_timeout"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	VendorID      types.String `tfsdk:"vendor_id"`
}

func (r *SourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (r *SourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP configuration source for an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The source ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The application ID this source belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The source name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The source type. Valid values: REST, GRAPHQL, MOCK, MCP_PROXY, FRONTEGG, CUSTOM_INTEGRATION.",
				Required:    true,
			},
			"source_url": schema.StringAttribute{
				Description: "The source URL (must be HTTPS).",
				Required:    true,
			},
			"api_timeout": schema.Int64Attribute{
				Description: "API timeout in milliseconds (500-5000). Defaults to 3000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3000),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the source is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"vendor_id": schema.StringAttribute{
				Description: "The vendor ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateSourceRequest{
		AppID:      data.ApplicationID.ValueString(),
		Name:       data.Name.ValueString(),
		Type:       data.Type.ValueString(),
		SourceURL:  data.SourceURL.ValueString(),
		APITimeout: int(data.APITimeout.ValueInt64()),
		Enabled:    data.Enabled.ValueBool(),
	}

	source, err := r.client.CreateSource(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create source: "+err.Error())
		return
	}

	data.ID = types.StringValue(source.ID)
	data.ApplicationID = types.StringValue(source.AppID)
	data.Name = types.StringValue(source.Name)
	data.Type = types.StringValue(source.Type)
	data.SourceURL = types.StringValue(source.SourceURL)
	data.APITimeout = types.Int64Value(int64(source.APITimeout))
	data.Enabled = types.BoolValue(source.Enabled)
	data.VendorID = types.StringValue(source.VendorID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.GetSourceByID(ctx, data.ApplicationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read source: "+err.Error())
		return
	}

	if source == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(source.ID)
	data.ApplicationID = types.StringValue(source.AppID)
	data.Name = types.StringValue(source.Name)
	data.Type = types.StringValue(source.Type)
	data.SourceURL = types.StringValue(source.SourceURL)
	data.APITimeout = types.Int64Value(int64(source.APITimeout))
	data.Enabled = types.BoolValue(source.Enabled)
	data.VendorID = types.StringValue(source.VendorID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := data.Enabled.ValueBool()
	updateReq := client.UpdateSourceRequest{
		AppID:      data.ApplicationID.ValueString(),
		Name:       data.Name.ValueString(),
		Type:       data.Type.ValueString(),
		SourceURL:  data.SourceURL.ValueString(),
		APITimeout: int(data.APITimeout.ValueInt64()),
		Enabled:    &enabled,
	}

	source, err := r.client.UpdateSource(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update source: "+err.Error())
		return
	}

	data.ID = types.StringValue(source.ID)
	data.ApplicationID = types.StringValue(source.AppID)
	data.Name = types.StringValue(source.Name)
	data.Type = types.StringValue(source.Type)
	data.SourceURL = types.StringValue(source.SourceURL)
	data.APITimeout = types.Int64Value(int64(source.APITimeout))
	data.Enabled = types.BoolValue(source.Enabled)
	data.VendorID = types.StringValue(source.VendorID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSource(ctx, data.ApplicationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete source: "+err.Error())
		return
	}
}

func (r *SourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: application_id:source_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'application_id:source_id'",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
