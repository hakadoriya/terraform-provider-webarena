// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kunitsucom/webarena-go/indigo"
)

// Ensure WebARENAProvider satisfies various provider interfaces.
var (
	_ provider.Provider = (*WebARENAProvider)(nil)
	// _ provider.ProviderWithFunctions = (*WebARENAProvider)(nil) // Commented out because it is not implemented.
)

// WebARENAProvider defines the provider implementation.
type WebARENAProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// WebARENAProviderModel describes the provider data model.
type WebARENAProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *WebARENAProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "webarena"
	resp.Version = p.version
}

func (p *WebARENAProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description:         "WebARENA API endpoint",
				MarkdownDescription: "WebARENA API endpoint",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				Description:         "WebARENA API client ID",
				MarkdownDescription: "WebARENA API client ID",
				Optional:            true,
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				Description:         "WebARENA API client secret",
				MarkdownDescription: "WebARENA API client secret",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *WebARENAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config WebARENAProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown WebARENA API Endpoint",
			"The provider cannot create the WebARENA API client as there is an unknown configuration value for the WebARENA API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the WEBARENA_INDIGO_ENDPOINT environment variable.",
		)
	}

	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown WebARENA API ClientID",
			"The provider cannot create the WebARENA API client as there is an unknown configuration value for the WebARENA API client_id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the WEBARENA_INDIGO_CLIENT_ID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown WebARENA API ClientSecret",
			"The provider cannot create the WebARENA API client as there is an unknown configuration value for the WebARENA API client_secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the WEBARENA_INDIGO_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	endpoint := os.Getenv(indigo.WEBARENA_INDIGO_ENDPOINT)
	clientID := os.Getenv(indigo.WEBARENA_INDIGO_CLIENT_ID)
	clientSecret := os.Getenv(indigo.WEBARENA_INDIGO_CLIENT_SECRET)

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueString()
	}
	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing WebARENA API ClientID",
			"The provider cannot create the WebARENA API client as there is a missing or empty value for the WebARENA API client_id. "+
				"Set the client_id value in the configuration or use the WEBARENA_INDIGO_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing WebARENA API ClientSecret",
			"The provider cannot create the WebARENA API client as there is a missing or empty value for the WebARENA API client_secret. "+
				"Set the client_secret value in the configuration or use the WEBARENA_INDIGO_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	opts := []indigo.ClientOption{
		indigo.ClientOptionWithClientID(clientID),
		indigo.ClientOptionWithClientSecret(clientSecret),
	}
	if endpoint != "" {
		opts = append(opts, indigo.ClientOptionWithEndpoint(endpoint))
	}

	// Create a new WebARENA client using the configuration values
	indigoClient, err := indigo.NewClient(ctx, opts...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create WebARENA API Client",
			"An unexpected error occurred when creating the WebARENA API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"WebARENA Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = indigoClient
	resp.ResourceData = indigoClient
}

func (p *WebARENAProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIndigoV1VmSSHKeyResource,
	}
}

func (p *WebARENAProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIndigoV1VmSSHKeyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WebARENAProvider{
			version: version,
		}
	}
}
