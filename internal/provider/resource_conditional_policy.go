package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-frontegg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ConditionalPolicyResource{}
var _ resource.ResourceWithImportState = &ConditionalPolicyResource{}

func NewConditionalPolicyResource() resource.Resource {
	return &ConditionalPolicyResource{}
}

// ConditionalPolicyResource defines the resource implementation.
type ConditionalPolicyResource struct {
	client *client.Client
}

// ConditionalPolicyResourceModel describes the resource data model.
type ConditionalPolicyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	AppIDs          types.List   `tfsdk:"app_ids"`
	TenantID        types.String `tfsdk:"tenant_id"`
	InternalToolIDs types.List   `tfsdk:"internal_tool_ids"`
	Targeting       types.Object `tfsdk:"targeting"`
	Metadata        types.Map    `tfsdk:"metadata"`
}

func (r *ConditionalPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conditional_policy"
}

func (r *ConditionalPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a conditional policy for access control with targeting rules.",
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
			"internal_tool_ids": schema.ListAttribute{
				Description: "List of internal tool IDs this policy applies to. Empty list applies to all tools.",
				Required:    true,
				ElementType: types.StringType,
			},
			"targeting": schema.SingleNestedAttribute{
				Description: "Targeting rules for when this policy applies.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"if": schema.SingleNestedAttribute{
						Description: "Conditions block.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"conditions": schema.ListNestedAttribute{
								Description: "List of conditions to evaluate.",
								Required:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"attribute": schema.StringAttribute{
											Description: "The attribute to evaluate.",
											Required:    true,
										},
										"negate": schema.BoolAttribute{
											Description: "Whether to negate the condition.",
											Required:    true,
										},
										"op": schema.StringAttribute{
											Description: "The operation to perform.",
											Required:    true,
										},
										"value": schema.MapAttribute{
											Description: "The value to compare against.",
											Required:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"then": schema.SingleNestedAttribute{
						Description: "Result block.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"result": schema.StringAttribute{
								Description: "The result when conditions are met. Valid values: ALLOW, DENY, APPROVAL_REQUIRED.",
								Required:    true,
							},
							"approval_flow_id": schema.StringAttribute{
								Description: "The approval flow ID (required when result is APPROVAL_REQUIRED).",
								Optional:    true,
							},
						},
					},
				},
			},
			"metadata": schema.MapAttribute{
				Description: "Additional metadata for the policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ConditionalPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConditionalPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConditionalPolicyResourceModel

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

	// Convert internal_tool_ids
	var toolIDs []string
	resp.Diagnostics.Append(data.InternalToolIDs.ElementsAs(ctx, &toolIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build targeting from state
	targeting := r.buildTargetingFromState(ctx, data.Targeting)

	// Convert metadata
	var metadata map[string]interface{}
	if !data.Metadata.IsNull() {
		metadata = make(map[string]interface{})
		var metadataStr map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataStr, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range metadataStr {
			metadata[k] = v
		}
	}

	createReq := client.CreateConditionalPolicyRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         data.Enabled.ValueBool(),
		AppIDs:          appIDs,
		TenantID:        data.TenantID.ValueString(),
		InternalToolIDs: toolIDs,
		Targeting:       targeting,
		Metadata:        metadata,
	}

	policy, err := r.client.CreateConditionalPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create conditional policy: "+err.Error())
		return
	}

	data.ID = types.StringValue(policy.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConditionalPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConditionalPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetConditionalPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read conditional policy: "+err.Error())
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

func (r *ConditionalPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConditionalPolicyResourceModel

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

	// Convert internal_tool_ids
	var toolIDs []string
	resp.Diagnostics.Append(data.InternalToolIDs.ElementsAs(ctx, &toolIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build targeting from state
	targeting := r.buildTargetingFromState(ctx, data.Targeting)

	// Convert metadata
	var metadata map[string]interface{}
	if !data.Metadata.IsNull() {
		metadata = make(map[string]interface{})
		var metadataStr map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadataStr, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range metadataStr {
			metadata[k] = v
		}
	}

	enabled := data.Enabled.ValueBool()
	updateReq := client.UpdateConditionalPolicyRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		Enabled:         &enabled,
		AppIDs:          appIDs,
		TenantID:        data.TenantID.ValueString(),
		InternalToolIDs: toolIDs,
		Targeting:       targeting,
		Metadata:        metadata,
	}

	_, err := r.client.UpdateConditionalPolicy(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update conditional policy: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConditionalPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConditionalPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete conditional policy: "+err.Error())
		return
	}
}

func (r *ConditionalPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ConditionalPolicyResource) buildTargetingFromState(ctx context.Context, targeting types.Object) *client.PolicyTargeting {
	if targeting.IsNull() || targeting.IsUnknown() {
		return nil
	}

	// This is a simplified implementation - in production you'd want to properly
	// parse the nested object structure
	return nil
}
