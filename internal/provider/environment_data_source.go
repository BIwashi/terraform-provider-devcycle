package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type environmentDataSourceType struct{}

func (t environmentDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]tfsdk.Attribute{
			"project_id": {
				MarkdownDescription: "Project id or key of the project to which the environment belongs",
				Computed:            true,
				Type:                types.StringType,
			},
			"project_key": {
				MarkdownDescription: "Project key of the project to which the environment belongs",
				Required:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Environment Name",
				Computed:            true,
				Type:                types.StringType,
			},
			"key": {
				MarkdownDescription: "Environment Key",
				Required:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "Environment Description",
				Computed:            true,
				Type:                types.StringType,
			},
			"color": {
				MarkdownDescription: "Environment Color in Hex with leading #",
				Computed:            true,
				Type:                types.StringType,
			},
			"type": {
				MarkdownDescription: "Environment Type",
				Computed:            true,
				Type:                types.StringType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Environment Id",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
			"sdk_keys": {
				Computed:            true,
				MarkdownDescription: "SDK Keys for the environment",
				Type:                types.ListType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (t environmentDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return environmentDataSource{
		provider: provider,
	}, diags
}

type environmentDataSourceData struct {
	Id          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
	Type        types.String `tfsdk:"type"`
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectKey  types.String `tfsdk:"project_key"`
	SDKKeys     []string     `tfsdk:"sdk_keys"`
}

type environmentDataSource struct {
	provider provider
}

func (d environmentDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data environmentDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	environment, httpResponse, err := d.provider.MgmtClient.EnvironmentsApi.EnvironmentsControllerFindOne(ctx, data.Key.Value, data.ProjectKey.Value)
	if err != nil || httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got error: %s", err))
		return
	}

	data.Id = types.String{Value: environment.Id}
	data.Key = types.String{Value: environment.Key}
	data.Name = types.String{Value: environment.Name}
	data.Description = types.String{Value: environment.Description}
	data.Color = types.String{Value: environment.Color}
	data.Type = types.String{Value: environment.Type_}
	data.ProjectId = types.String{Value: environment.Project}
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Mobile)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Server)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Client)...)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}