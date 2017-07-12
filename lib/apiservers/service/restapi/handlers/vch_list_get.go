package handlers

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/vic/cmd/vic-machine/common"
	"github.com/vmware/vic/lib/apiservers/service/models"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/util"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/lib/install/data"
	"github.com/vmware/vic/lib/install/management"
	"github.com/vmware/vic/lib/install/validate"
	"github.com/vmware/vic/pkg/version"
	"github.com/vmware/vic/pkg/vsphere/vm"
)

// VCHListGet is the handler for listing VCHs
type VCHListGet struct {
}

// VCHListGet is the handler for listing VCHs within a Datacenter
type VCHDatacenterListGet struct {
}

func (h *VCHListGet) Handle(params operations.GetTargetVchParams, principal interface{}) middleware.Responder {
	d := buildData(
		url.URL{Host: params.Target},
		principal.(Credentials).user,
		principal.(Credentials).pass,
		params.Thumbprint,
		params.ComputeResource)

	vchs, err := handle(d)
	if err != nil {
		return operations.NewGetTargetVchDefault(err.Code()).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewGetTargetVchOK().WithPayload(operations.GetTargetVchOKBody{Vchs: vchs})
}

func (h *VCHDatacenterListGet) Handle(params operations.GetTargetDatacenterDatacenterVchParams, principal interface{}) middleware.Responder {
	d := buildData(
		url.URL{Host: params.Target},
		principal.(Credentials).user,
		principal.(Credentials).pass,
		params.Thumbprint,
		params.ComputeResource)
	// TODO: include datacenter!

	vchs, err := handle(d)
	if err != nil {
		return operations.NewGetTargetVchDefault(err.Code()).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewGetTargetVchOK().WithPayload(operations.GetTargetVchOKBody{Vchs: vchs})
}

func handle(d *data.Data) ([]*models.VCHListItem, *util.HttpError) {
	validator, err := validateTarget(d)
	if err != nil {
		return nil, util.NewHttpError(400, err.Error())
	}

	executor := management.NewDispatcher(validator.Context, validator.Session, nil, false)
	vchs, err := executor.SearchVCHs(validator.ClusterPath)
	if err != nil {
		return nil, util.NewHttpError(500, fmt.Sprintf("Failed to search VCHs in %s: %s", validator.ResourcePoolPath, err))
	}

	return vchsToModels(vchs, executor), nil
}

func buildData(url url.URL, user string, pass string, thumbprint *string, computeResource *string) *data.Data {
	d := data.Data{
		Target: &common.Target{
			URL:      &url,
			User:     user,
			Password: &pass,
		},
	}

	if thumbprint != nil {
		d.Thumbprint = *thumbprint
	}

	if computeResource != nil {
		d.ComputeResourcePath = *computeResource
	}

	return &d
}

func validateTarget(d *data.Data) (*validate.Validator, error) {
	if err := d.HasCredentials(); err != nil {
		return nil, fmt.Errorf("Invalid Credentials: %s", err)
	}

	ctx := context.Background()

	validator, err := validate.NewValidator(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Validation Error: %s", err)
	}
	// If dc is not set, and multiple datacenter is available, vic-machine ls will list VCHs under all datacenters.
	validator.AllowEmptyDC()

	_, err = validator.ValidateTarget(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Target validation failed: %s", err)
	}
	_, err = validator.ValidateCompute(ctx, d, false)
	if err != nil {
		return nil, fmt.Errorf("Compute resource validation failed: %s", err)
	}

	return validator, nil
}

// Copied from list.go. TODO: deduplicate
func upgradeStatusMessage(ctx context.Context, vch *vm.VirtualMachine, installerVer *version.Build, vchVer *version.Build) string {
	if sameVer := installerVer.Equal(vchVer); sameVer {
		return "Up to date"
	}

	upgrading, err := vch.VCHUpdateStatus(ctx)
	if err != nil {
		return fmt.Sprintf("Unknown: %s", err)
	}
	if upgrading {
		return "Upgrade in progress"
	}

	canUpgrade, err := installerVer.IsNewer(vchVer)
	if err != nil {
		return fmt.Sprintf("Unknown: %s", err)
	}
	if canUpgrade {
		return fmt.Sprintf("Upgradeable to %s", installerVer.ShortVersion())
	}

	oldInstaller, err := installerVer.IsOlder(vchVer)
	if err != nil {
		return fmt.Sprintf("Unknown: %s", err)
	}
	if oldInstaller {
		return fmt.Sprintf("VCH has newer version")
	}

	// can't get here
	return "Invalid upgrade status"
}

func vchsToModels(vchs []*vm.VirtualMachine, executor *management.Dispatcher) []*models.VCHListItem {
	installerVer := version.GetBuild()
	payload := make([]*models.VCHListItem, 0)
	for _, vch := range vchs {
		var version *version.Build
		if vchConfig, err := executor.GetNoSecretVCHConfig(vch); err == nil {
			version = vchConfig.Version
		}

		parentPath := path.Dir(path.Dir(vch.InventoryPath))
		name := path.Base(vch.InventoryPath)
		upgradeStatus := upgradeStatusMessage(context.Background(), vch, installerVer, version)

		payload = append(payload, &models.VCHListItem{ID: vch.Reference().Value, Name: name, Path: parentPath, Version: version.ShortVersion(), UpgradeStatus: upgradeStatus})
	}

	return payload
}
