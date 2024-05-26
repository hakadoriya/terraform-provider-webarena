// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// MEMO: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-create
// MEMO: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-update
// MEMO: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-delete
// MEMO: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-import
// MEMO: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-acceptance-testing

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kunitsucom/webarena-go/indigo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = (*IndigoV1VmSSHKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*IndigoV1VmSSHKeyResource)(nil)
	_ resource.ResourceWithImportState = (*IndigoV1VmSSHKeyResource)(nil)
)

func NewIndigoV1VmSSHKeyResource() resource.Resource {
	return &IndigoV1VmSSHKeyResource{}
}

// IndigoV1VmSSHKeyResource defines the resource implementation.
type IndigoV1VmSSHKeyResource struct {
	client *indigo.Client
}

// IndigoV1VmSSHKeyResourceModel describes the resource data model.
type IndigoV1VmSSHKeyResourceModel struct {
	ID        types.String `tfsdk:"id"` // NOTE: WebARENA API uses id as int64, but Terraform import uses id as string.
	ServiceId types.String `tfsdk:"service_id"`
	UserId    types.Int64  `tfsdk:"user_id"`
	Name      types.String `tfsdk:"name"`
	SshKey    types.String `tfsdk:"sshkey"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *IndigoV1VmSSHKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + indigoV1VmSSHKeyName
}

func (r *IndigoV1VmSSHKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "WebARENA Indigo V1 VM SSH Key resource",
		MarkdownDescription: "WebARENA Indigo V1 VM SSH Key resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the SSH key. This is a unique number assigned by the provider.",
			},
			"service_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service to which the SSH key belongs.",
			},
			"user_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the user who owns the SSH key.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SSH key.",
			},
			"sshkey": schema.StringAttribute{
				Required:    true,
				Description: "The SSH public key. Example: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDQ...",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the SSH key. Possible values: `ACTIVE`, `DEACTIVE`. (`DEACTIVE` is not typo, it's a valid status defined by the provider.) Default: `ACTIVE`",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the SSH key was created. The date and time are in the UTC timezone and in the `YYYY-mm-dd HH:MM:SS` format.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the SSH key was last updated. The date and time are in the UTC timezone and in the `YYYY-mm-dd HH:MM:SS` format.",
			},
		},
	}
}

func (r *IndigoV1VmSSHKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*indigo.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *indigo.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *IndigoV1VmSSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IndigoV1VmSSHKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	createResp, err := r.client.CreateWebArenaIndigoV1VmSSHKey(ctx, &indigo.CreateWebArenaIndigoV1VmSSHKeyRequest{
		SshName: plan.Name.ValueString(),
		SshKey:  plan.SshKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sshkey, got error: %s", err))
		return
	}

	// Update the model with the response data
	plan.ID = types.StringValue(strconv.FormatInt(createResp.SshKey.Id, 10))
	plan.ServiceId = types.StringValue(createResp.SshKey.ServiceId)
	plan.Name = types.StringValue(createResp.SshKey.Name)
	plan.SshKey = types.StringValue(createResp.SshKey.Sshkey)
	plan.UserId = types.Int64Value(createResp.SshKey.UserId)
	plan.Status = types.StringValue(createResp.SshKey.Status)
	plan.CreatedAt = types.StringValue(createResp.SshKey.CreatedAt)
	plan.UpdatedAt = types.StringValue(createResp.SshKey.UpdatedAt)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("created a resource: id=%s", plan.ID.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IndigoV1VmSSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IndigoV1VmSSHKeyResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("ID Error", fmt.Sprintf("Unable to parse ID=%s, got error: %s", state.ID.ValueString(), err))
		return
	}

	retrieveResp, err := r.client.RetrieveWebArenaIndigoV1VmSSHKey(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sshkey, got error: %s", err))
		return
	}
	if retrieveResp == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read sshkey, got nil response")
		return
	}
	// must be 1
	if len(retrieveResp.SshKey) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("sshKey length is not 1: sshkey=%+v", retrieveResp.SshKey))
		return
	}

	sshKey := retrieveResp.SshKey[0]
	state.ID = types.StringValue(strconv.FormatInt(sshKey.Id, 10))
	state.ServiceId = types.StringValue(sshKey.ServiceId)
	state.UserId = types.Int64Value(sshKey.UserId)
	state.Name = types.StringValue(sshKey.Name)
	state.SshKey = types.StringValue(sshKey.Sshkey)
	state.Status = types.StringValue(sshKey.Status)
	state.CreatedAt = types.StringValue(sshKey.CreatedAt)
	state.UpdatedAt = types.StringValue(sshKey.UpdatedAt)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("read a resource: id=%s", state.ID.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *IndigoV1VmSSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IndigoV1VmSSHKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("ID Error", fmt.Sprintf("Unable to parse ID=%s, got error: %s", plan.ID.ValueString(), err))
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	updateResp, err := r.client.UpdateWebArenaIndigoV1VmSSHKey(ctx, id, &indigo.UpdateWebArenaIndigoV1VmSSHKeyRequest{
		SshName:     plan.Name.ValueString(),
		SshKey:      plan.SshKey.ValueString(),
		SshKeyState: plan.Status.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sshkey, got error: %s", err))
		return
	}
	if updateResp == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read sshkey, got nil response")
		return
	}

	retrieveResp, err := r.client.RetrieveWebArenaIndigoV1VmSSHKey(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sshkey, got error: %s", err))
		return
	}
	if retrieveResp == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read sshkey, got nil response")
		return
	}
	// must be 1
	if len(retrieveResp.SshKey) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("sshKey length is not 1: sshkey=%+v", retrieveResp.SshKey))
		return
	}

	sshKey := retrieveResp.SshKey[0]
	plan.ID = types.StringValue(strconv.FormatInt(sshKey.Id, 10))
	plan.ServiceId = types.StringValue(sshKey.ServiceId)
	plan.UserId = types.Int64Value(sshKey.UserId)
	plan.Name = types.StringValue(sshKey.Name)
	plan.SshKey = types.StringValue(sshKey.Sshkey)
	plan.Status = types.StringValue(sshKey.Status)
	plan.CreatedAt = types.StringValue(sshKey.CreatedAt)
	plan.UpdatedAt = types.StringValue(sshKey.UpdatedAt)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("updated a resource: id=%s", plan.ID.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IndigoV1VmSSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IndigoV1VmSSHKeyResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("ID Error", fmt.Sprintf("Unable to parse ID=%s, got error: %s", state.ID.ValueString(), err))
		return
	}

	destroyResp, err := r.client.DestroyWebArenaIndigoV1VmSSHKey(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sshkey, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a resource: success=%t msg=%s", destroyResp.Success, destroyResp.Message))
}

func (r *IndigoV1VmSSHKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
