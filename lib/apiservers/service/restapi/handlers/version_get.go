package handlers

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/pkg/version"
)

// VersionGet is the handler for accessing the version information for the service
type VersionGet struct {
}

func (h *VersionGet) Handle(params operations.GetVersionParams) middleware.Responder {
	return operations.NewGetVersionOK().WithPayload(version.GetBuild().ShortVersion())
}
