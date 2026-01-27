package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-agentlink/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IdentityConfigurationResource{}

func NewIdentityConfigurationResource() resource.Resource {
	return &IdentityConfigurationResource{}
}

// IdentityConfigurationResource defines the resource implementation.
type IdentityConfigurationResource struct {
	client *client.Client
}

// IdentityConfigurationResourceModel describes the resource data model.
type IdentityConfigurationResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	DefaultTokenExpiration types.Int64  `tfsdk:"default_token_expiration"`
}

func (r *IdentityConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_configuration"
}

func (r *IdentityConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages identity configuration settings including default token expiration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The configuration ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_token_expiration": schema.Int64Attribute{
				Description: "The default token expiration time in seconds. Minimum value is 10.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IdentityConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentityConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentityConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenExpiration := int(data.DefaultTokenExpiration.ValueInt64())

	updateReq := client.UpdateIdentityConfigurationRequest{
		DefaultTokenExpiration: &tokenExpiration,
	}

	config, err := r.client.UpdateIdentityConfiguration(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create identity configuration: "+err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(config.ID)
	data.DefaultTokenExpiration = types.Int64Value(int64(config.DefaultTokenExpiration))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentityConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetIdentityConfiguration(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read identity configuration: "+err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(config.ID)
	data.DefaultTokenExpiration = types.Int64Value(int64(config.DefaultTokenExpiration))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IdentityConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenExpiration := int(data.DefaultTokenExpiration.ValueInt64())

	updateReq := client.UpdateIdentityConfigurationRequest{
		DefaultTokenExpiration: &tokenExpiration,
	}

	config, err := r.client.UpdateIdentityConfiguration(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update identity configuration: "+err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(config.ID)
	data.DefaultTokenExpiration = types.Int64Value(int64(config.DefaultTokenExpiration))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Identity configuration cannot be deleted, it's a singleton resource.
	// On destroy, we simply remove it from state. The configuration will remain
	// on the server with its current values.
}
