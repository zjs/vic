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
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/vic/lib/apiservers/service/models"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/decode"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/errors"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/lib/install/data"
	"github.com/vmware/vic/lib/install/management"
	"github.com/vmware/vic/lib/install/validate"
	"github.com/vmware/vic/pkg/trace"
	"github.com/vmware/vic/pkg/version"
)

// VCHDelete is the handler for deleting a VCH
type VCHDelete struct {
}

// VCHDatacenterDelete is the handler for deleting a VCH within a Datacenter
type VCHDatacenterDelete struct {
}

func (h *VCHDelete) Handle(params operations.DeleteTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
	op := trace.FromContext(params.HTTPRequest.Context(), "VCHDelete: %s", params.VchID)

	b := buildDataParams{
		target:     params.Target,
		thumbprint: params.Thumbprint,
		vchID:      &params.VchID,
	}

	d, validator, err := buildDataAndValidateTarget(op, b, principal)
	if err != nil {
		return operations.NewDeleteTargetTargetVchVchIDDefault(errors.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	err = deleteVCH(op, d, validator, params.DeletionSpecification)
	if err != nil {
		return operations.NewDeleteTargetTargetVchVchIDDefault(errors.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewDeleteTargetTargetVchVchIDAccepted()
}

func (h *VCHDatacenterDelete) Handle(params operations.DeleteTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
	op := trace.FromContext(params.HTTPRequest.Context(), "VCHDelete: %s", params.VchID)

	b := buildDataParams{
		target:     params.Target,
		thumbprint: params.Thumbprint,
		datacenter: &params.Datacenter,
		vchID:      &params.VchID,
	}

	d, validator, err := buildDataAndValidateTarget(op, b, principal)
	if err != nil {
		return operations.NewDeleteTargetTargetDatacenterDatacenterVchVchIDDefault(errors.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	err = deleteVCH(op, d, validator, params.DeletionSpecification)
	if err != nil {
		return operations.NewDeleteTargetTargetDatacenterDatacenterVchVchIDDefault(errors.StatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewDeleteTargetTargetDatacenterDatacenterVchVchIDAccepted()
}

func deleteVCH(op trace.Operation, d *data.Data, validator *validate.Validator, specification *models.DeletionSpecification) error {
	executor := management.NewDispatcher(op, validator.Session(), management.ActionDelete, false)
	vch, err := executor.NewVCHFromID(d.ID)
	if err != nil {
		return errors.NewError(http.StatusNotFound, "failed to find VCH: %s", err)
	}

	err = validator.SetDataFromVM(op, vch, d)
	if err != nil {
		return errors.NewError(http.StatusInternalServerError, "failed to load VCH data: %s", err)
	}

	vchConfig, err := executor.GetNoSecretVCHConfig(vch)
	if err != nil {
		return errors.NewError(http.StatusInternalServerError, "failed to load VCH data: %s", err)
	}

	// compare vch version and vic-machine version
	installerBuild := version.GetBuild()
	if vchConfig.Version == nil || !installerBuild.Equal(vchConfig.Version) {
		op.Debugf("VCH version %q is different than API version %s", vchConfig.Version.ShortVersion(), installerBuild.ShortVersion())
	}

	deleteContainers, deleteVolumeStores := decode.FromDeletionSpecification(specification)
	err = executor.DeleteVCH(vchConfig, deleteContainers, deleteVolumeStores)
	if err != nil {
		return errors.NewError(http.StatusInternalServerError, "failed to delete VCH: %s", err)
	}

	return nil
}
