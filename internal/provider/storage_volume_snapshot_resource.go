package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeSnapshotResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeSnapshotResource is a helper function to simplify the provider implementation.
func NewStorageVolumeSnapshotResource() resource.Resource {
	return &StorageVolumeSnapshotResource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_snapshot_resource",
		},
	}
}

// StorageVolumeSnapshotResource defines the resource implementation.
type StorageVolumeSnapshotResource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeSnapshotResourceModel describes the resource data model.
type StorageVolumeSnapshotResourceModel struct {
	CxProfileName      types.String      `tfsdk:"cx_profile_name"`
	Name               types.String      `tfsdk:"name"`
	Volume             NameResourceModel `tfsdk:"volume"`
	SVM                NameResourceModel `tfsdk:"svm"`
	ExpiryTime         types.String      `tfsdk:"expiry_time"`
	SnaplockExpiryTime types.String      `tfsdk:"snaplock_expiry_time"`
	Comment            types.String      `tfsdk:"comment"`
	SnapmirrorLabel    types.String      `tfsdk:"snapmirror_label"`
	ID                 types.String      `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *StorageVolumeSnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Storage Volume Snapshot resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Snapshot name",
				Required:            true,
			},
			"volume": schema.SingleNestedAttribute{
				MarkdownDescription: "Volume the snapshot is on",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Volume Name",
						Required:            true,
					},
				},
			},
			"svm": schema.SingleNestedAttribute{
				MarkdownDescription: "svm the snapshot is on",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "svm Name",
						Required:            true,
					},
				},
			},
			"expiry_time": schema.StringAttribute{
				MarkdownDescription: "Snapshot copies with an expiry time set are not allowed to be deleted until the retetion time is reached",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment",
				Optional:            true,
			},
			"snapmirror_label": schema.StringAttribute{
				MarkdownDescription: "Label for SnapMirror Operations",
				Optional:            true,
			},
			"snaplock_expiry_time": schema.StringAttribute{
				MarkdownDescription: "Expiry time for Snapshot copy locking enabled volumes",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "storage/volumes/snapshots identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *StorageVolumeSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeSnapshotResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.StorageVolumeSnapshotResourceModel

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVM.Name.ValueString())
	if err != nil {
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No svm found", fmt.Sprintf("svm %s not found.", data.SVM.Name))
		return
	}
	volume, err := interfaces.GetUUIDVolumeByName(errorHandler, *client, svm.UUID, data.Volume.Name.ValueString())
	if err != nil {
		return
	}
	if volume == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("volume %s not found.", data.Volume.Name))
		return
	}

	request.Name = data.Name.ValueString()
	if !data.ExpiryTime.IsNull() {
		request.ExpiryTime = data.ExpiryTime.ValueString()
	}
	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
	}
	if !data.SnapmirrorLabel.IsNull() {
		request.SnapmirrorLabel = data.SnapmirrorLabel.ValueString()
	}
	if !data.SnaplockExpiryTime.IsNull() {
		request.SnaplockExpiryTime = data.SnaplockExpiryTime.ValueString()
	}

	snapshot, err := interfaces.CreateStorageVolumeSnapshot(errorHandler, *client, request, volume.UUID)
	if err != nil {
		return
	}
	// TODO: add async calls or add wait condition for create
	data.ID = types.StringValue(snapshot.UUID)
	tflog.Trace(ctx, "created a resource")
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageVolumeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageVolumeSnapshotResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVM.Name.ValueString())
	if err != nil {
		return
	}
	volume, err := interfaces.GetUUIDVolumeByName(errorHandler, *client, svm.UUID, data.Volume.Name.ValueString())
	if err != nil {
		return
	}
	snapshot, err := interfaces.GetStorageVolumeSnapshot(errorHandler, *client, volume.UUID, data.ID.ValueString())
	if err != nil {
		return
	}
	data.Name = types.StringValue(snapshot.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageVolumeSnapshotResourceModel
	var state *StorageVolumeSnapshotResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVM.Name.ValueString())
	if err != nil {
		return
	}
	volume, err := interfaces.GetUUIDVolumeByName(errorHandler, *client, svm.UUID, data.Volume.Name.ValueString())
	if err != nil {
		return
	}
	var request interfaces.StorageVolumeSnapshotResourceModel
	if !data.Name.Equal(state.Name) {
		// rename snapshot
		request.Name = data.Name.ValueString()
	}
	if !data.ExpiryTime.Equal(state.ExpiryTime) {
		if data.ExpiryTime.ValueString() == "" {
			errorHandler.MakeAndReportError("update expiry_time", "expiry_time cannot be updated with empty string")
			return
		}
		request.ExpiryTime = data.ExpiryTime.ValueString()
	}
	if !data.SnaplockExpiryTime.Equal(state.SnaplockExpiryTime) {
		if data.SnaplockExpiryTime.ValueString() == "" {
			errorHandler.MakeAndReportError("update snaplock_expiry_time", "snaplock_expiry_time cannot be updated with empty string")
			return
		}
		request.SnaplockExpiryTime = data.SnaplockExpiryTime.ValueString()
	}
	if !data.Comment.Equal(state.Comment) {
		if data.Comment.ValueString() == "" {
			errorHandler.MakeAndReportError("update comment", "comment cannot be updated with empty string")
			return
		}
		request.Comment = data.Comment.ValueString()
	}
	if !data.SnapmirrorLabel.Equal(state.SnapmirrorLabel) {
		if data.SnapmirrorLabel.ValueString() == "" {
			errorHandler.MakeAndReportError("update snapmirror_label", "snapmirror_label cannot be updated with empty string")
			return
		}
		request.SnapmirrorLabel = data.SnapmirrorLabel.ValueString()
	}
	tflog.Debug(ctx, fmt.Sprintf("update a resource %s: %#v", state.ID.ValueString(), request))
	err = interfaces.UpdateStorageVolumeSnapshot(errorHandler, *client, request, volume.UUID, state.ID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeSnapshotResourceModel
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
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVM.Name.ValueString())
	if err != nil {
		return
	}
	volume, err := interfaces.GetUUIDVolumeByName(errorHandler, *client, svm.UUID, data.Volume.Name.ValueString())
	if err != nil {
		return
	}
	err = interfaces.DeleteStorageVolumeSnapshot(errorHandler, *client, volume.UUID, data.ID.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}