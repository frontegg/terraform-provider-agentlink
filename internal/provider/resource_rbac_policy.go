package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-agentlink/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RbacPolicyResource{}
var _ resource.ResourceWithImportState = &RbacPolicyResource{}

func NewRbacPolicyResource() resource.Resource {
	return &RbacPolicyResource{}
}

// RbacPolicyResource defines the resource implementation.
type RbacPolicyResource struct {
	client *client.Client
}

// RbacPolicyResourceModel describes the resource data model.
type RbacPolicyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	AppIDs          types.List   `tfsdk:"app_ids"`
	TenantID        types.String `tfsdk:"tenant_id"`
	Type            types.String `tfsdk:"type"`
	Keys            types.List   `tfsdk:"keys"`
	InternalToolIDs types.List   `tfsdk:"internal_tool_ids"`
}

func (r *RbacPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rbac_policy"
}

func (r *RbacPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an RBAC (Role-Based Access Control) policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The policy ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The policy name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The policy description.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is enabled.",
				Required:    true,
			},
			"app_ids": schema.ListAttribute{
				Description: "List of application IDs this policy applies to.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"tenant_id": schema.StringAttribute{
				Description: "The tenant ID this policy applies to.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The RBAC policy type. Valid values: RBAC_ROLES, RBAC_PERMISSIONS.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"keys": schema.ListAttribute{
				Description: "List of role or permission keys.",
				Required:    true,
				ElementType: types.StringType,
			},
			"internal_tool_ids": schema.ListAttribute{
				Description: "List of internal tool IDs this policy applies to. At least one is required.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *RbacPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RbacPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RbacPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert app_ids
	var appIDs []string
	if !data.AppIDs.IsNull() {
		resp.Diagnostics.Append(data.AppIDs.ElementsAs(ctx, &appIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert keys
	var keys []string
	resp.Diagnostics.Append(data.Keys.ElementsAs(ctx, &keys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert internal_tool_ids
	var toolIDs []string
	resp.Diagnostics.Append(data.InternalToolIDs.ElementsAs(ctx, &toolIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(toolIDs) == 0 {
		resp.Diagnostics.AddError("Validation Error", "At least one internal_tool_id is required for RBAC policies")
		return
	}

	createReq := client.CreateRbacPolicyRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		AppIDs:          appIDs,
		TenantID:        data.TenantID.ValueString(),
		Type:            data.Type.ValueString(),
		Keys:            keys,
		InternalToolIDs: toolIDs,
	}

	policy, err := r.client.CreateRbacPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create RBAC policy: "+err.Error())
		return
	}

	data.ID = types.StringValue(policy.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RbacPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RbacPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetRbacPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read RBAC policy: "+err.Error())
		return
	}

	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Description = types.StringValue(policy.Description)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.Type = types.StringValue(policy.Type)

	// Convert app_ids
	if len(policy.AppIDs) > 0 {
		appIDValues := make([]attr.Value, len(policy.AppIDs))
		for i, id := range policy.AppIDs {
			appIDValues[i] = types.StringValue(id)
		}
		data.AppIDs, _ = types.ListValue(types.StringType, appIDValues)
	} else {
		data.AppIDs = types.ListNull(types.StringType)
	}

	if policy.TenantID != "" {
		data.TenantID = types.StringValue(policy.TenantID)
	}

	// Convert keys
	if len(policy.Keys) > 0 {
		keyValues := make([]attr.Value, len(policy.Keys))
		for i, key := range policy.Keys {
			keyValues[i] = types.StringValue(key)
		}
		data.Keys, _ = types.ListValue(types.StringType, keyValues)
	} else {
		data.Keys, _ = types.ListValue(types.StringType, []attr.Value{})
	}

	// Convert internal_tool_ids
	if len(policy.InternalToolIDs) > 0 {
		toolIDValues := make([]attr.Value, len(policy.InternalToolIDs))
		for i, id := range policy.InternalToolIDs {
			toolIDValues[i] = types.StringValue(id)
		}
		data.InternalToolIDs, _ = types.ListValue(types.StringType, toolIDValues)
	} else {
		data.InternalToolIDs, _ = types.ListValue(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RbacPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RbacPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert app_ids
	var appIDs []string
	if !data.AppIDs.IsNull() {
		resp.Diagnostics.Append(data.AppIDs.ElementsAs(ctx, &appIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert keys
	var keys []string
	resp.Diagnostics.Append(data.Keys.ElementsAs(ctx, &keys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert internal_tool_ids
	var toolIDs []string
	resp.Diagnostics.Append(data.InternalToolIDs.ElementsAs(ctx, &toolIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := data.Enabled.ValueBool()
	updateReq := client.UpdateRbacPolicyRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         &enabled,
		AppIDs:          appIDs,
		TenantID:        data.TenantID.ValueString(),
		Keys:            keys,
		InternalToolIDs: toolIDs,
	}

	_, err := r.client.UpdateRbacPolicy(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update RBAC policy: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RbacPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RbacPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete RBAC policy: "+err.Error())
		return
	}
}

func (r *RbacPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
