package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ZiconProvider satisfies the provider.Provider interface.
var _ provider.Provider = &ZiconProvider{}

// ZiconProvider is the Terraform provider implementation for ZiCON Cloud.
type ZiconProvider struct {
	// version is set by main.go via NewProvider, used for the user agent
	// and left empty during local (dev) builds.
	version string
}

// ZiconProviderModel maps the provider configuration schema to Go types.
type ZiconProviderModel struct {
	AccessToken types.String `tfsdk:"access_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZiconProvider{version: version}
	}
}

func (p *ZiconProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zicon"
	resp.Version = p.version
}

func (p *ZiconProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interacts with the ZiCON Cloud API.",
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Supabase session token used to authenticate project-creation requests against the ZiCON Cloud API.",
			},
		},
	}
}

func (p *ZiconProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ZiconProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown ZiCON Access Token",
			"The provider cannot create the ZiCON API client because the access_token attribute is unknown.",
		)
		return
	}

	if config.AccessToken.IsNull() || config.AccessToken.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing ZiCON Access Token",
			"The provider requires a non-empty access_token to authenticate against the ZiCON Cloud API.",
		)
		return
	}

	client := NewClient(config.AccessToken.ValueString())

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *ZiconProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
	}
}

func (p *ZiconProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
