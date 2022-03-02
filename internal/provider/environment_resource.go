package provider

import (
	"context"
	"fmt"
	devcyclem "github.com/devcyclehq/go-mgmt-sdk"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type environmentResourceType struct{}

func (t environmentResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]tfsdk.Attribute{
			"project_id": {
				MarkdownDescription: "Project id or key of the project to which the environment belongs",
				Required:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Environment Name",
				Required:            true,
				Type:                types.StringType,
			},
			"key": {
				MarkdownDescription: "Environment Key",
				Required:            true,
				Type:                types.StringType,
			},
			"description": {
				MarkdownDescription: "Environment Description",
				Required:            true,
				Type:                types.StringType,
			},
			"color": {
				MarkdownDescription: "Environment Color in Hex with leading #",
				Required:            true,
				Type:                types.StringType,
			},
			"type": {
				MarkdownDescription: "Environment Type",
				Required:            true,
				Type:                types.StringType,
			},
			"appiconuri": {
				MarkdownDescription: "Environment App Icon URI",
				Required:            true,
				Type:                types.StringType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Environment Id",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"sdkKeys": {
				Computed:            true,
				MarkdownDescription: "SDK Keys for the environment",
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"type": {
						Type:     types.StringType,
						Computed: true,
					},
					"key": {
						Type:     types.StringType,
						Computed: true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (t environmentResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return environmentResource{
		provider: provider,
	}, diags
}

type environmentResourceData struct {
	Id          types.String                    `tfsdk:"id"`
	Key         types.String                    `tfsdk:"key"`
	Name        types.String                    `tfsdk:"name"`
	Description types.String                    `tfsdk:"description"`
	Color       types.String                    `tfsdk:"color"`
	Type        types.String                    `tfsdk:"type"`
	Settings    environmentResourceDataSettings `tfsdk:"settings"`
	ProjectId   types.String                    `tfsdk:"project_id"`
	SDKKeys     []string                        `tfsdk:"sdkKeys"`
}

func sdkKeyConvert(keys []devcyclem.ApiKey) []string {
	var sdkKeys []string
	for _, sdkKey := range keys {
		sdkKeys = append(sdkKeys, sdkKey.Key)
	}
	return sdkKeys
}

type environmentResourceDataSettings struct {
	AppIconURI types.String `tfsdk:"appiconuri"`
}

func (s *environmentResourceDataSettings) toCreateSDK() *devcyclem.AllOfCreateEnvironmentDtoSettings {
	return &devcyclem.AllOfCreateEnvironmentDtoSettings{
		AppIconURI: s.AppIconURI.Value,
	}
}
func (s *environmentResourceDataSettings) toUpdateSDK() *devcyclem.AllOfUpdateEnvironmentDtoSettings {
	return &devcyclem.AllOfUpdateEnvironmentDtoSettings{
		AppIconURI: s.AppIconURI.Value,
	}
}

type environmentResource struct {
	provider provider
}

func (r environmentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data environmentResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, httpResponse, err := r.provider.MgmtClient.EnvironmentsApi.EnvironmentsControllerCreate(ctx, devcyclem.CreateEnvironmentDto{
		Name:        data.Name.Value,
		Key:         data.Key.Value,
		Description: data.Description.Value,
		Color:       data.Color.Value,
		Type_:       data.Type.Value,
		Settings:    data.Settings.toCreateSDK(),
	}, data.ProjectId.Value)
	if err != nil || httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}

	data.Id.Value = environment.Id
	data.Key.Value = environment.Key
	data.Name.Value = environment.Name
	data.Description.Value = environment.Description
	data.Color.Value = environment.Color
	data.Type.Value = environment.Type_
	data.ProjectId.Value = environment.Project
	data.Settings.AppIconURI.Value = environment.Settings.AppIconURI
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Mobile)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Server)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Client)...)

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r environmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data environmentResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, httpResponse, err := r.provider.MgmtClient.EnvironmentsApi.EnvironmentsControllerFindOne(ctx, data.Key.Value, data.ProjectId.Value)
	if err != nil || httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got error: %s", err))
		return
	}

	data.Id.Value = environment.Id
	data.Key.Value = environment.Key
	data.Name.Value = environment.Name
	data.Description.Value = environment.Description
	data.Color.Value = environment.Color
	data.Type.Value = environment.Type_
	data.ProjectId.Value = environment.Project
	data.Settings.AppIconURI.Value = environment.Settings.AppIconURI
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Mobile)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Server)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Client)...)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r environmentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data environmentResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, httpResponse, err := r.provider.MgmtClient.EnvironmentsApi.EnvironmentsControllerUpdate(ctx, devcyclem.UpdateEnvironmentDto{
		Name:        data.Name.Value,
		Key:         data.Key.Value,
		Description: data.Description.Value,
		Color:       data.Color.Value,
		Type_:       data.Type.Value,
		Settings:    data.Settings.toUpdateSDK(),
	}, data.Key.Value, data.ProjectId.Value)
	if err != nil || httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment, got error: %s", err))
		return
	}

	data.Id.Value = environment.Id
	data.Key.Value = environment.Key
	data.Name.Value = environment.Name
	data.Description.Value = environment.Description
	data.Color.Value = environment.Color
	data.Type.Value = environment.Type_
	data.ProjectId.Value = environment.Project
	data.Settings.AppIconURI.Value = environment.Settings.AppIconURI
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Mobile)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Server)...)
	data.SDKKeys = append(data.SDKKeys, sdkKeyConvert(environment.SdkKeys.Client)...)

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r environmentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data environmentResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.provider.MgmtClient.EnvironmentsApi.EnvironmentsControllerRemove(ctx, data.Key.Value, data.ProjectId.Value)
	if err != nil || httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r environmentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
