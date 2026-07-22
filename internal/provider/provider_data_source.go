package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// providerCountsDSAttrs returns the computed attributes for the counts
// nested object in the data source schema.
func providerCountsDSAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"customers": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Number of customers under this provider.",
		},
		"locations": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Number of locations under this provider.",
		},
		"connectors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Number of connectors under this provider.",
		},
	}
}

var _ datasource.DataSource = &ccProviderDataSource{}

func NewCCProviderDataSource() datasource.DataSource {
	return &ccProviderDataSource{}
}

type ccProviderDataSource struct {
	client *customerConnectData
}

func (d *ccProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (d *ccProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single NetFoundry Customer Connect Provider by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the provider to look up.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the provider.",
			},
			"network_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network identifier associated with this provider.",
			},
			"organization_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Organisation identifier associated with this provider.",
			},
			"network_group_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network Group identifier associated with this provider.",
			},
			"internal_customer_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal customer that holds locations attached directly to this provider.",
			},
			"owner_identity_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity that owns this provider.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity that created this provider.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when this provider was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when this provider was last updated.",
			},
			"deleted_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when this provider was deleted.",
			},
			"deleted_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identity that deleted this provider.",
			},
			"deleted": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this provider has been deleted.",
			},
			"counts": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Sub-resource counts for this provider.",
				Attributes:          providerCountsDSAttrs(),
			},
		},
	}
}

func (d *ccProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring provider data source")
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*customerConnectData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *customerConnectData, got %T", req.ProviderData))
		return
	}
	d.client = data
}

func (d *ccProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state providerEntityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading provider data source", map[string]any{"id": state.ID.ValueString()})

	url := fmt.Sprintf("%s/providers/%s", d.client.apiBaseURL, state.ID.ValueString())

	respBody, _, err := doRequest(ctx, http.MethodGet, url, d.client.accessToken, nil)
	if err != nil {
		if errors.Is(err, errNotFound) {
			resp.Diagnostics.AddError("Provider not found",
				fmt.Sprintf("No provider found with id %q", state.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Read provider failed", err.Error())
		return
	}

	var p apiProviderEntity
	if err := json.Unmarshal(respBody, &p); err != nil {
		resp.Diagnostics.AddError("Parse error", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, providerEntityFromAPI(p))...)
}
