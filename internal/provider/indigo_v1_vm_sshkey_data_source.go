// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kunitsucom/webarena-go/indigo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &IndigoV1VmSSHKeyDataSource{}

func NewIndigoV1VmSSHKeyDataSource() datasource.DataSource {
	return &IndigoV1VmSSHKeyDataSource{}
}

// IndigoV1VmSSHKeyDataSource defines the data source implementation.
type IndigoV1VmSSHKeyDataSource struct {
	client *indigo.Client
}

// IndigoV1VmSSHKeyDataSourceModel describes the data source data model.
type IndigoV1VmSSHKeyDataSourceModel struct {
	Id        types.Int64  `tfsdk:"id"`
	ServiceId types.String `tfsdk:"service_id"`
	UserId    types.Int64  `tfsdk:"user_id"`
	Name      types.String `tfsdk:"name"`
	SshKey    types.String `tfsdk:"sshkey"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *IndigoV1VmSSHKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + indigoV1VmSSHKeyName
}

func (d *IndigoV1VmSSHKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "WebARENA Indigo V1 VM SSH Key data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
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
				Computed:    true,
				Description: "The name of the SSH key.",
			},
			"sshkey": schema.StringAttribute{
				Computed:    true,
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

func (d *IndigoV1VmSSHKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*indigo.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *indigo.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *IndigoV1VmSSHKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IndigoV1VmSSHKeyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	retrieveResp, err := d.client.RetrieveWebArenaIndigoV1VmSSHKey(ctx, state.Id.ValueInt64())
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Sshkey length is not 1: sshkey=%v", retrieveResp.SshKey))
		return
	}

	sshKey := retrieveResp.SshKey[0]
	state.Id = types.Int64Value(sshKey.Id)
	state.ServiceId = types.StringValue(sshKey.ServiceId)
	state.UserId = types.Int64Value(sshKey.UserId)
	state.Name = types.StringValue(sshKey.Name)
	state.SshKey = types.StringValue(sshKey.Sshkey)
	state.Status = types.StringValue(sshKey.Status)
	state.CreatedAt = types.StringValue(sshKey.CreatedAt)
	state.UpdatedAt = types.StringValue(sshKey.UpdatedAt)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("read a data source: id=%d", state.Id.ValueInt64()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
