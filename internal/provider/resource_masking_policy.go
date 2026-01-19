package provider

import (
	"context"

	"github.com/frontegg/terraform-provider-frontegg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MaskingPolicyResource{}
var _ resource.ResourceWithImportState = &MaskingPolicyResource{}

func NewMaskingPolicyResource() resource.Resource {
	return &MaskingPolicyResource{}
}

// MaskingPolicyResource defines the resource implementation.
type MaskingPolicyResource struct {
	client *client.Client
}

// MaskingPolicyResourceModel describes the resource data model.
type MaskingPolicyResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	AppIDs              types.List   `tfsdk:"app_ids"`
	TenantID            types.String `tfsdk:"tenant_id"`
	InternalToolIDs     types.List   `tfsdk:"internal_tool_ids"`
	PolicyConfiguration types.Object `tfsdk:"policy_configuration"`
}

// MaskingConfigModel represents the masking configuration
type MaskingConfigModel struct {
	CreditCard      types.Bool `tfsdk:"credit_card"`
	EmailAddress    types.Bool `tfsdk:"email_address"`
	PhoneNumber     types.Bool `tfsdk:"phone_number"`
	IpAddress       types.Bool `tfsdk:"ip_address"`
	UsSsn           types.Bool `tfsdk:"us_ssn"`
	UsDriverLicense types.Bool `tfsdk:"us_driver_license"`
	UsPassport      types.Bool `tfsdk:"us_passport"`
	UsItin          types.Bool `tfsdk:"us_itin"`
	UsBankNumber    types.Bool `tfsdk:"us_bank_number"`
	IbanCode        types.Bool `tfsdk:"iban_code"`
	SwiftCode       types.Bool `tfsdk:"swift_code"`
	BitcoinAddress  types.Bool `tfsdk:"bitcoin_address"`
	EthereumAddress types.Bool `tfsdk:"ethereum_address"`
	CvvCvc          types.Bool `tfsdk:"cvv_cvc"`
	Url             types.Bool `tfsdk:"url"`
}

func (r *MaskingPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_masking_policy"
}

func (r *MaskingPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a data masking policy for sensitive information protection.",
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
			"policy_configuration": schema.SingleNestedAttribute{
				Description: "Configuration specifying what data types to mask.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"credit_card": schema.BoolAttribute{
						Description: "Enable credit card detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"email_address": schema.BoolAttribute{
						Description: "Enable email address detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"phone_number": schema.BoolAttribute{
						Description: "Enable phone number detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ip_address": schema.BoolAttribute{
						Description: "Enable IP address detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"us_ssn": schema.BoolAttribute{
						Description: "Enable US Social Security Number detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"us_driver_license": schema.BoolAttribute{
						Description: "Enable US driver license detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"us_passport": schema.BoolAttribute{
						Description: "Enable US passport detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"us_itin": schema.BoolAttribute{
						Description: "Enable US ITIN detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"us_bank_number": schema.BoolAttribute{
						Description: "Enable US bank number detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"iban_code": schema.BoolAttribute{
						Description: "Enable IBAN code detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"swift_code": schema.BoolAttribute{
						Description: "Enable SWIFT code detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"bitcoin_address": schema.BoolAttribute{
						Description: "Enable Bitcoin address detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ethereum_address": schema.BoolAttribute{
						Description: "Enable Ethereum address detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"cvv_cvc": schema.BoolAttribute{
						Description: "Enable CVV/CVC detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"url": schema.BoolAttribute{
						Description: "Enable URL detection and masking.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}

func (r *MaskingPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MaskingPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MaskingPolicyResourceModel

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

	// Parse policy configuration
	policyConfig := r.extractPolicyConfig(ctx, data.PolicyConfiguration, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateMaskingPolicyRequest{
		Name:                data.Name.ValueString(),
		Description:         data.Description.ValueString(),
		Enabled:             data.Enabled.ValueBool(),
		AppIDs:              appIDs,
		TenantID:            data.TenantID.ValueString(),
		InternalToolIDs:     toolIDs,
		PolicyConfiguration: policyConfig,
	}

	policy, err := r.client.CreateMaskingPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create masking policy: "+err.Error())
		return
	}

	data.ID = types.StringValue(policy.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaskingPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MaskingPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetMaskingPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read masking policy: "+err.Error())
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

func (r *MaskingPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MaskingPolicyResourceModel

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

	// Parse policy configuration
	policyConfig := r.extractPolicyConfig(ctx, data.PolicyConfiguration, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := data.Enabled.ValueBool()
	updateReq := client.UpdateMaskingPolicyRequest{
		Name:                data.Name.ValueString(),
		Description:         data.Description.ValueString(),
		Enabled:             &enabled,
		AppIDs:              appIDs,
		TenantID:            data.TenantID.ValueString(),
		InternalToolIDs:     toolIDs,
		PolicyConfiguration: policyConfig,
	}

	_, err := r.client.UpdateMaskingPolicy(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update masking policy: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaskingPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MaskingPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete masking policy: "+err.Error())
		return
	}
}

func (r *MaskingPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MaskingPolicyResource) extractPolicyConfig(ctx context.Context, configObj types.Object, diags *diag.Diagnostics) *client.MaskingPolicyConfiguration {
	if configObj.IsNull() || configObj.IsUnknown() {
		return nil
	}

	attrs := configObj.Attributes()

	getBool := func(key string) bool {
		if v, ok := attrs[key]; ok {
			if boolVal, ok := v.(types.Bool); ok {
				return boolVal.ValueBool()
			}
		}
		return false
	}

	return &client.MaskingPolicyConfiguration{
		CreditCard:      getBool("credit_card"),
		EmailAddress:    getBool("email_address"),
		PhoneNumber:     getBool("phone_number"),
		IpAddress:       getBool("ip_address"),
		UsSsn:           getBool("us_ssn"),
		UsDriverLicense: getBool("us_driver_license"),
		UsPassport:      getBool("us_passport"),
		UsItin:          getBool("us_itin"),
		UsBankNumber:    getBool("us_bank_number"),
		IbanCode:        getBool("iban_code"),
		SwiftCode:       getBool("swift_code"),
		BitcoinAddress:  getBool("bitcoin_address"),
		EthereumAddress: getBool("ethereum_address"),
		CvvCvc:          getBool("cvv_cvc"),
		Url:             getBool("url"),
	}
}
