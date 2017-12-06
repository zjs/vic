// Copyright 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/vmware/vic/cmd/vic-machine/common"
	"github.com/vmware/vic/lib/apiservers/service/models"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/util"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/lib/install/data"
	"github.com/vmware/vic/lib/install/management"
	"github.com/vmware/vic/lib/install/validate"
	"github.com/vmware/vic/pkg/trace"
	"github.com/vmware/vic/pkg/vsphere/vm"
)

// VCHAction is the handler for performing an action on a VCH
type VCHAction struct {
}

// VCHDatacenterAction is the handler for performing an action on a VCH within a Datacenter
type VCHDatacenterAction struct {
}

// action represents a valid params.Action value (as those are not currently captured in an enum)
type action string

const (
	Debug action    = "debug"
	Rollback action = "rollback"
	Upgrade action  = "upgrade"
)

func (h *VCHAction) Handle(params operations.PostTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
	op := trace.FromContext(params.HTTPRequest.Context(), "VCHAction: %s (%s)", params.VchID, params.Action)

	b := buildDataParams{
		target:     params.Target,
		thumbprint: params.Thumbprint,
	}

	d, err := buildData(op, b, principal)
	if err != nil {
		return operations.NewPostTargetTargetVchVchIDDefault(util.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	d.ID = params.VchID

	task, err := actOnVch(op, d, params.Action)
	if err != nil {
		return operations.NewPostTargetTargetVchVchIDDefault(util.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewPostTargetTargetVchVchIDAccepted().WithPayload(operations.PostTargetTargetVchVchIDAcceptedBody{Task: task})
}

func (h *VCHDatacenterAction) Handle(params operations.PostTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
	op := trace.FromContext(params.HTTPRequest.Context(), "VCHDatacenterAction: %s (%s)", params.VchID, params.Action)

	b := buildDataParams{
		target:     params.Target,
		thumbprint: params.Thumbprint,
	}

	d, err := buildData(op, b, principal)
	if err != nil {
		return operations.NewPostTargetTargetDatacenterDatacenterVchVchIDDefault(util.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	d.ID = params.VchID

	task, err := actOnVch(op, d, params.Action)
	if err != nil {
		return operations.NewPostTargetTargetDatacenterDatacenterVchVchIDDefault(util.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewPostTargetTargetDatacenterDatacenterVchVchIDAccepted().WithPayload(operations.PostTargetTargetDatacenterDatacenterVchVchIDAcceptedBody{Task: task})
}

func actionFromString(action string) (action, error) {
	switch action {
	case string(Debug):
		return Debug, nil
	case string(Rollback):
		return Rollback, nil
	case string(Upgrade):
		return Upgrade, nil
	default:
		return "", util.WrapError(http.StatusUnprocessableEntity, fmt.Errorf("Unknown action: %s", action))
	}
}

func actOnVch(op trace.Operation, d *data.Data, action string) (*strfmt.URI, error) {
	a, err := actionFromString(action)
	if err != nil {
		return nil, err
	}

	validator, err := validateTarget(op, d)
	if err != nil {
		return nil, util.WrapError(http.StatusBadRequest, err)
	}

	executor := management.NewDispatcher(validator.Context, validator.Session, nil, false)
	vch, err := executor.NewVCHFromID(d.ID)
	if err != nil {
		return nil, util.NewError(http.StatusNotFound, fmt.Sprintf("Failed to find VCH: %s", err))
	}

	var task *strfmt.URI
	switch a {
	case Debug:
		task, err = debugVch(op, vch)
	case Upgrade:
		d.Rollback = false
		task, err = upgradeVch(op, validator, executor, d, vch)
	case Rollback:
		d.Rollback = true
		task, err = upgradeVch(op, validator, executor, d, vch)
	}

	if err != nil {
		op.Error(err)
	}

	return task, err
}

func debugVch(op trace.Operation, vch *vm.VirtualMachine) (*strfmt.URI, error) {
	return nil, util.NewError(http.StatusNotImplemented, "Debug not implemented")
}

func upgradeVch(op trace.Operation, validator *validate.Validator, executor *management.Dispatcher, d *data.Data, vch *vm.VirtualMachine) (*strfmt.URI, error) {
	upgrading, err := vch.VCHUpdateStatus(op)
	if err != nil {
		return nil, util.WrapError(http.StatusConflict, fmt.Errorf("Unable to determine if upgrade/configure is in progress: %s", err))
	}
	if upgrading {
		return nil, util.WrapError(http.StatusConflict, fmt.Errorf("Another upgrade/configure operation is in progress"))
	}

	if err = vch.SetVCHUpdateStatus(op, true); err != nil {
		return nil, util.WrapError(http.StatusInternalServerError, fmt.Errorf("Failed to set UpdateInProgress flag to true: %s", err))
	}

	defer func() {
		if err = vch.SetVCHUpdateStatus(op, false); err != nil {
			op.Errorf("Failed to reset UpdateInProgress: %s", err)
		}
	}()

	vchConfig, err := executor.FetchAndMigrateVCHConfig(vch)
	if err != nil {
		return nil, util.WrapError(http.StatusInternalServerError, fmt.Errorf("Failed to get Virtual Container Host configuration: %s", err))
	}

	vConfig := validator.AddDeprecatedFields(op, vchConfig, d)

	images := common.Images{}
	vConfig.ImageFiles, err = images.CheckImagesFiles(op, true)
	vConfig.ApplianceISO = path.Base(images.ApplianceISO)
	vConfig.BootstrapISO = path.Base(images.BootstrapISO)
	vConfig.Timeout = time.Hour // FIXME

	// only care about versions if we're not doing a manual rollback
	if !d.Rollback {
		if err := validator.AssertVersionForAPI(op, vchConfig); err != nil {
			return nil, util.WrapError(http.StatusInternalServerError, fmt.Errorf("Version check failed: %s", err))
		}
	}

	if vchConfig, err = validator.ValidateMigratedConfig(op, vchConfig); err != nil {
		return nil, util.WrapError(http.StatusInternalServerError, fmt.Errorf("Failed to migrate Virtual Container Host configuration: %s", err))
	}

	if !d.Rollback {
		err = executor.Configure(vch, vchConfig, vConfig, false)
	} else {
		err = executor.Rollback(vch, vchConfig, vConfig)
	}

	if err != nil {
		return nil, util.WrapError(http.StatusInternalServerError, err)
	}

	return nil, nil
}
