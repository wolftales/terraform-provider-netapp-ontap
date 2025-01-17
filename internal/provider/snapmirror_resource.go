package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SnapmirrorResource{}
var _ resource.ResourceWithImportState = &SnapmirrorResource{}

// NewSnapmirrorResource is a helper function to simplify the provider implementation.
func NewSnapmirrorResource() resource.Resource {
	return &SnapmirrorResource{
		config: resourceOrDataSourceConfig{
			name: "snapmirror_resource",
		},
	}
}

// SnapmirrorResource defines the resource implementation.
type SnapmirrorResource struct {
	config resourceOrDataSourceConfig
}

// SnapmirrorResourceModel describes the resource data model.
type SnapmirrorResourceModel struct {
	CxProfileName       types.String       `tfsdk:"cx_profile_name"`
	SourceEndPoint      *EndPoint          `tfsdk:"source_endpoint"`
	DestinationEndPoint *EndPoint          `tfsdk:"destination_endpoint"`
	CreateDestination   *CreateDestination `tfsdk:"create_destination"`
	Initialize          types.Bool         `tfsdk:"initialize"`
	Healthy             types.Bool         `tfsdk:"healthy"`
	State               types.String       `tfsdk:"state"`
	ID                  types.String       `tfsdk:"id"`
}

// EndPoint describes source/destination endpoint data model.
type EndPoint struct {
	Cluster *Cluster     `tfsdk:"cluster"`
	Path    types.String `tfsdk:"path"`
}

// CreateDestination describes CreateDestination data model.
type CreateDestination struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// Cluster describes Cluster data model.
type Cluster struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name
func (r *SnapmirrorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *SnapmirrorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Snapmirror resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"source_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror source endpoint",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cluster": schema.SingleNestedAttribute{
						MarkdownDescription: "Cluster details",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "cluster name",
								Required:            true,
							},
						},
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Path to the source endpoint of the SnapMirror relationship",
						Required:            true,
					},
				},
			},
			"destination_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror destination endpoint",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cluster": schema.SingleNestedAttribute{
						MarkdownDescription: "Cluster details",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "cluster name",
								Required:            true,
							},
						},
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Path to the destination endpoint of the SnapMirror relationship",
						Required:            true,
					},
				},
			},
			"create_destination": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror privision destination",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable this property to provision the destination endpoint",
						Required:            true,
					},
				},
			},
			"initialize": schema.BoolAttribute{
				MarkdownDescription: "initialize the relationship",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"healthy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SnapmirrorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please resport this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SnapmirrorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapmirrorResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
		return
	}

	restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSnapmirrorByID
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.Healthy = types.BoolValue(restInfo.Healthy)
	data.State = types.StringValue(restInfo.State)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a snapmirror resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SnapmirrorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SnapmirrorResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SnapmirrorResourceBodyDataModelONTAP

	if resp.Diagnostics.HasError() {
		return
	}

	body.SourceEndPoint.Path = data.SourceEndPoint.Path.ValueString()
	body.DestinationEndPoint.Path = data.DestinationEndPoint.Path.ValueString()
	if data.SourceEndPoint.Cluster != nil {
		if !data.SourceEndPoint.Cluster.Name.IsNull() {
			body.SourceEndPoint.Cluster.Name = data.SourceEndPoint.Cluster.Name.ValueString()
		}
	}
	if data.DestinationEndPoint.Cluster != nil {
		if !data.DestinationEndPoint.Cluster.Name.IsNull() {
			body.DestinationEndPoint.Cluster.Name = data.DestinationEndPoint.Cluster.Name.ValueString()
		}
	}
	if data.CreateDestination != nil {
		if !data.CreateDestination.Enabled.IsNull() {
			body.CreateDestination.Enabled = data.CreateDestination.Enabled.ValueBool()
		}
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSnapmirror(errorHandler, *client, body)
	if err != nil {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("create snapmirror resource: %#v", resource))

	restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSnapmirror
		return
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror info: %#v", restInfo))
	data.Healthy = types.BoolValue(restInfo.Healthy)
	data.State = types.StringValue(restInfo.State)
	data.ID = types.StringValue(resource.UUID)

	if data.Initialize.ValueBool() && data.State.ValueString() == "uninitialized" {
		time.Sleep(3 * time.Second)
		err := interfaces.InitializeSnapmirror(errorHandler, *client, data.ID.ValueString(), "snapmirrored")
		if err != nil {
			// error reporting done inside InitializeSnapmirror
			return
		}
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror info: %#v", restInfo))
		data.Healthy = types.BoolValue(restInfo.Healthy)
		data.State = types.StringValue(restInfo.State)
	}
	restInfo, err = interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSnapmirror
		return
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror info: %#v", restInfo))
	// Update the computed parameters
	data.Healthy = types.BoolValue(restInfo.Healthy)
	data.State = types.StringValue(restInfo.State)
	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a snapmirror resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SnapmirrorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SnapmirrorResourceModel
	var state SnapmirrorResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// License updates are not supported
	err := errorHandler.MakeAndReportError("Update not supported for snapmirror", "Update not supported for snapmirror")
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SnapmirrorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SnapmirrorResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "snapmirror UUID is null")
		return
	}

	err = interfaces.DeleteSnapmirror(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SnapmirrorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
