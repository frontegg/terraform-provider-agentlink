package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-frontegg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource defines the resource implementation.
type ApplicationResource struct {
	client *client.Client
}

// ApplicationResourceModel describes the resource data model.
type ApplicationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	VendorID      types.String `tfsdk:"vendor_id"`
	Name          types.String `tfsdk:"name"`
	AppURL        types.String `tfsdk:"app_url"`
	LoginURL      types.String `tfsdk:"login_url"`
	LogoURL       types.String `tfsdk:"logo_url"`
	AccessType    types.String `tfsdk:"access_type"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	Type          types.String `tfsdk:"type"`
	FrontendStack types.String `tfsdk:"frontend_stack"`
	Description   types.String `tfsdk:"description"`
	AllowDcr      types.Bool   `tfsdk:"allow_dcr"`
	AppHost       types.String `tfsdk:"app_host"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frontegg application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The application ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vendor_id": schema.StringAttribute{
				Description: "The vendor ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The application name.",
				Required:    true,
			},
			"app_url": schema.StringAttribute{
				Description: "The application URL.",
				Required:    true,
			},
			"login_url": schema.StringAttribute{
				Description: "The login URL.",
				Required:    true,
			},
			"logo_url": schema.StringAttribute{
				Description: "The logo URL.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"access_type": schema.StringAttribute{
				Description: "The access type. Valid values: FREE_ACCESS, MANAGED_ACCESS.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("FREE_ACCESS"),
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default application.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the application is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The application type. Valid values: web, mobile-ios, mobile-android, agent, other.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("agent"),
			},
			"frontend_stack": schema.StringAttribute{
				Description: "The frontend stack. Valid values: react, angular, vue, nextjs, other.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("react"),
			},
			"description": schema.StringAttribute{
				Description: "The application description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"allow_dcr": schema.BoolAttribute{
				Description: "Whether to allow Dynamic Client Registration (DCR).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"app_host": schema.StringAttribute{
				Description: "The application host (computed by Frontegg).",
				Computed:    true,
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create with all planned values
	isDefault := data.IsDefault.ValueBool()
	isActive := data.IsActive.ValueBool()
	allowDcr := data.AllowDcr.ValueBool()

	createReq := client.CreateApplicationRequest{
		Name:          data.Name.ValueString(),
		AppURL:        data.AppURL.ValueString(),
		LoginURL:      data.LoginURL.ValueString(),
		LogoURL:       data.LogoURL.ValueString(),
		AccessType:    data.AccessType.ValueString(),
		IsDefault:     &isDefault,
		IsActive:      &isActive,
		Type:          data.Type.ValueString(),
		FrontendStack: data.FrontendStack.ValueString(),
		Description:   data.Description.ValueString(),
		AllowDcr:      &allowDcr,
	}

	app, err := r.client.CreateApplication(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create application: "+err.Error())
		return
	}

	// Map response to model
	r.mapApplicationToModel(app, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplicationByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read application: "+err.Error())
		return
	}

	if app == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to model
	r.mapApplicationToModel(app, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isDefault := data.IsDefault.ValueBool()
	isActive := data.IsActive.ValueBool()
	allowDcr := data.AllowDcr.ValueBool()

	updateReq := client.UpdateApplicationRequest{
		Name:        data.Name.ValueString(),
		AppURL:      data.AppURL.ValueString(),
		LoginURL:    data.LoginURL.ValueString(),
		LogoURL:     data.LogoURL.ValueString(),
		AccessType:  data.AccessType.ValueString(),
		IsDefault:   &isDefault,
		IsActive:    &isActive,
		Type:        data.Type.ValueString(),
		Description: data.Description.ValueString(),
		AllowDcr:    &allowDcr,
	}

	app, err := r.client.UpdateApplication(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update application: "+err.Error())
		return
	}

	// Map response to model
	r.mapApplicationToModel(app, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete application: "+err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapApplicationToModel maps an Application response to the resource model
func (r *ApplicationResource) mapApplicationToModel(app *client.Application, data *ApplicationResourceModel) {
	data.ID = types.StringValue(app.ID)
	data.VendorID = types.StringValue(app.VendorID)
	data.Name = types.StringValue(app.Name)
	data.AppURL = types.StringValue(app.AppURL)
	data.LoginURL = types.StringValue(app.LoginURL)
	data.LogoURL = types.StringValue(app.LogoURL)
	data.AccessType = types.StringValue(app.AccessType)
	data.IsDefault = types.BoolValue(app.IsDefault)
	data.IsActive = types.BoolValue(app.IsActive)
	data.Type = types.StringValue(app.Type)
	data.FrontendStack = types.StringValue(app.FrontendStack)
	data.Description = types.StringValue(app.Description)
	data.AllowDcr = types.BoolValue(app.AllowDcr)

	// AppHost may be empty if not set by Frontegg
	if app.AppHost != "" {
		data.AppHost = types.StringValue(app.AppHost)
	} else {
		data.AppHost = types.StringValue("")
	}
}
