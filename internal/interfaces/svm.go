package interfaces

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SvmGetDataModelONTAP describes the GET record data model using go types for mapping.
type SvmGetDataModelONTAP struct {
	Name string
	UUID string
}

// SvmDataModelONTAP describes the svm info required by other API's request.
type SvmDataModelONTAP struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	Name           string              `mapstructure:"name"`
	Ipspace        Ipspace             `mapstructure:"ipspace"`
	SnapshotPolicy SnapshotPolicy      `mapstructure:"snapshot_policy"`
	SubType        string              `mapstructure:"subtype,omitempty"`
	Comment        string              `mapstructure:"comment,omitempty"`
	Language       string              `mapstructure:"language,omitempty"`
	MaxVolumes     string              `mapstructure:"max_volumes,omitempty"`
	Aggregates     []map[string]string `mapstructure:"aggregates,omitempty"`
}

// Ipspace describes the resource data model.
type Ipspace struct {
	Name string `mapstructure:"name,omitempty"`
}

// SnapshotPolicy describes the resource data model.
type SnapshotPolicy struct {
	Name string `mapstructure:"name,omitempty"`
}

// GetSvm to get svm info by uuid
func GetSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*SvmGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading vserver info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP *SvmGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read vserver info: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetSvmByName to get svm info by name
func GetSvmByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*SvmGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading vserver info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP *SvmGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm by name", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read vserver info: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSvm to create vserver
func CreateSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, data SvmResourceModel) (*SvmGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding vserver body", fmt.Sprintf("error on encoding svm/svms body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("svm/svms", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating vserver", fmt.Sprintf("error on POST svm/svms: %s, statusCode %d", err, statusCode))

	}

	var dataONTAP SvmGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from POST svm/svms", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create vserver source - udata: %#v", dataONTAP))
	return &dataONTAP, nil

}

// DeleteSvm to delete vserver
func DeleteSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "svm/svms/" + uuid
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting vserver", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))

	}
	return nil
}

// UpdateSvm to update a vserver
func UpdateSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, data SvmResourceModel, uuid string, rename bool) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding vserver body", fmt.Sprintf("error on encoding svm/svms body: %s, body: %#v", err, data))
	}
	// Name is only passed to patch if it is a rename
	if !rename {
		delete(body, "name")
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("svm/svms/"+uuid, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating vserver", fmt.Sprintf("error on PATCH svm/svms: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// ValidateIntORString to validate int or string
func ValidateIntORString(errorHandler *utils.ErrorHandler, value string, astring string) error {
	if value == "" || value == astring {
		return nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return errorHandler.MakeAndReportError("falied to validate", fmt.Sprintf("Error: expecting int value or '%s', got: %s", astring, value))
	}

	stringValue := strconv.Itoa(intValue)
	if stringValue != value {
		return errorHandler.MakeAndReportError("falied to validate", fmt.Sprintf("Error: expecting int value or '%s', got: %s", astring, value))
	}
	return nil
}