package provider

import (
	"context"
	"strings"

	"github.com/frontegg/terraform-provider-agentlink/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// normalizeOrigin removes trailing slashes from URLs for consistent comparison
func normalizeOrigin(url string) string {
	return strings.TrimSuffix(url, "/")
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AllowedOriginsResource{}
var _ resource.ResourceWithImportState = &AllowedOriginsResource{}

func NewAllowedOriginsResource() resource.Resource {
	return &AllowedOriginsResource{}
}

// AllowedOriginsResource defines the resource implementation.
type AllowedOriginsResource struct {
	client *client.Client
}

// AllowedOriginsResourceModel describes the resource data model.
type AllowedOriginsResourceModel struct {
	ID             types.String `tfsdk:"id"`
	AllowedOrigins types.Set    `tfsdk:"allowed_origins"`
}

func (r *AllowedOriginsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowed_origins"
}

func (r *AllowedOriginsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the allowed origins (CORS) configuration for the Frontegg vendor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The vendor ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_origins": schema.SetAttribute{
				Description: "Set of allowed origins for CORS. These URLs are permitted to make requests to the Frontegg API.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *AllowedOriginsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AllowedOriginsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AllowedOriginsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract allowed origins from the list
	var origins []string
	resp.Diagnostics.Append(data.AllowedOrigins.ElementsAs(ctx, &origins, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.UpdateAllowedOrigins(ctx, origins)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update allowed origins: "+err.Error())
		return
	}

	data.ID = types.StringValue(config.ID)
	// Keep the planned allowed_origins - the API may normalize URLs (add trailing slashes)
	// but we want to keep the user's original values in state

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AllowedOriginsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AllowedOriginsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetVendorConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read vendor config: "+err.Error())
		return
	}

	data.ID = types.StringValue(config.ID)

	// Normalize the allowed origins from the response (remove trailing slashes)
	normalizedOrigins := make([]string, len(config.AllowedOrigins))
	for i, origin := range config.AllowedOrigins {
		normalizedOrigins[i] = normalizeOrigin(origin)
	}

	allowedOriginsSet, diags := types.SetValueFrom(ctx, types.StringType, normalizedOrigins)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.AllowedOrigins = allowedOriginsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AllowedOriginsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AllowedOriginsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract allowed origins from the list
	var origins []string
	resp.Diagnostics.Append(data.AllowedOrigins.ElementsAs(ctx, &origins, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.UpdateAllowedOrigins(ctx, origins)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update allowed origins: "+err.Error())
		return
	}

	data.ID = types.StringValue(config.ID)
	// Keep the planned allowed_origins - the API may normalize URLs (add trailing slashes)
	// but we want to keep the user's original values in state

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AllowedOriginsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// On delete, we set allowed origins to an empty list
	// This effectively removes all custom allowed origins
	_, err := r.client.UpdateAllowedOrigins(ctx, []string{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to clear allowed origins: "+err.Error())
		return
	}
}

func (r *AllowedOriginsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For import, we just read the current state from the API
	config, err := r.client.GetVendorConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read vendor config: "+err.Error())
		return
	}

	var data AllowedOriginsResourceModel
	data.ID = types.StringValue(config.ID)

	// Normalize the allowed origins (remove trailing slashes)
	normalizedOrigins := make([]string, len(config.AllowedOrigins))
	for i, origin := range config.AllowedOrigins {
		normalizedOrigins[i] = normalizeOrigin(origin)
	}

	allowedOriginsSet, diags := types.SetValueFrom(ctx, types.StringType, normalizedOrigins)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.AllowedOrigins = allowedOriginsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
