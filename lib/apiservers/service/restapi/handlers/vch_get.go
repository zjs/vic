package handlers

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"

	"github.com/docker/docker/opts"
	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware/vic/lib/apiservers/service/models"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/util"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/lib/config/executor"
	"github.com/vmware/vic/lib/install/data"
	"github.com/vmware/vic/lib/install/management"
	"github.com/vmware/vic/pkg/version"
	"github.com/vmware/vic/pkg/vsphere/vm"
	"github.com/vmware/vic/lib/install/validate"
)

// VCHGet is the handler for inspecting a VCH
type VCHGet struct {
}

func (h *VCHGet) Handle(params operations.GetTargetVchVchIDParams, principal interface{}) middleware.Responder {
	d := buildData(
		url.URL{Host: params.Target},
		principal.(Credentials).user,
		principal.(Credentials).pass,
		params.Thumbprint,
		nil)

	d.ID = params.VchID

	vch, err := handle2(d)

	if err != nil {
		return operations.NewGetTargetVchVchIDDefault(err.Code()).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewGetTargetVchVchIDOK().WithPayload(vch)
}

func handle2(d *data.Data) (*models.VCH, *util.HttpError) { // TODO: naming is hard when scope is broad.
	validator, err := validateTarget(d)
	if err != nil {
		return nil, util.NewHttpError(400, err.Error())
	}

	executor := management.NewDispatcher(validator.Context, validator.Session, nil, false)
	vch, err := executor.NewVCHFromID(d.ID)
	if err != nil {
		return nil, util.NewHttpError(500, fmt.Sprintf("Failed to inspect VCH: %s", err))
	}

	err = validate.SetDataFromVM(validator.Context, validator.Session.Finder, vch, d)
	if err != nil {
		return nil, util.NewHttpError(500, fmt.Sprintf("Failed to load VCH data: %s", err))
	}

	return vchToModel(vch, d, executor), nil
}

func vchToModel(vch *vm.VirtualMachine, d *data.Data, executor *management.Dispatcher) *models.VCH {
	vchConfig, err := executor.GetNoSecretVCHConfig(vch)
	if err != nil {
		// TODO
	}

	model := &models.VCH{}
	model.Version = models.Version(vchConfig.Version.ShortVersion())
	model.Name = vchConfig.Name

	// compute
	model.Compute = &models.VCHCompute{
		CPU: &models.VCHComputeCPU{
			Limit:       asMHz(&vchConfig.Container.ContainerVMSize.CPU.Limit),
			Reservation: asMHz(&vchConfig.Container.ContainerVMSize.CPU.Reservation),
		},
		Memory: &models.VCHComputeMemory{
			Limit:       asBytes(&vchConfig.Container.ContainerVMSize.Memory.Limit),
			Reservation: asBytes(&vchConfig.Container.ContainerVMSize.Memory.Reservation),
		},
		Resource: &models.ManagedObject{
			ID: vchConfig.Container.ComputeResources[0].String(), // TODO: This doesn't seem to be an ID?
		},
	}

	if vchConfig.Container.ContainerVMSize.CPU.Shares != nil {
		model.Compute.CPU.Shares = &models.Shares{
			Level:  string(vchConfig.Container.ContainerVMSize.CPU.Shares.Level),
			Number: int64(vchConfig.Container.ContainerVMSize.CPU.Shares.Shares),
		}
	}

	if vchConfig.Container.ContainerVMSize.Memory.Shares != nil {
		model.Compute.Memory.Shares = &models.Shares{
			Level:  string(vchConfig.Container.ContainerVMSize.Memory.Shares.Level),
			Number: int64(vchConfig.Container.ContainerVMSize.Memory.Shares.Shares),
		}
	}

	// network
	model.Network = &models.VCHNetwork{
		Bridge: &models.VCHNetworkBridge{
			PortGroup: &models.ManagedObject{
				ID: vchConfig.Network.BridgeNetwork,
			},
			IPRange: asIPRange(vchConfig.Network.BridgeIPRange),
		},
		Client:     asNetwork(vchConfig.Networks["client"]),
		Management: asNetwork(vchConfig.Networks["management"]),
		Public:     asNetwork(vchConfig.Networks["public"]),
	}

	containerNetworks := make([]*models.ContainerNetwork, 0, len(vchConfig.Network.ContainerNetworks))
	for key, value := range vchConfig.Network.ContainerNetworks {
		if key != "bridge" {
			containerNetworks = append(containerNetworks, &models.ContainerNetwork{
				Alias: value.Name,
				PortGroup: &models.ManagedObject{
					ID: value.Common.ID, // TODO: This also doesn't seem to be an ID.
				},
				Nameservers: *asIPAddresses(&value.Nameservers),
				Gateway: &models.Gateway{
					Address:             asIPAddress(value.Gateway.IP),
					RoutingDestinations: []models.IPRange{asIPRange(&value.Gateway)},
				},
				IPRanges: *asIPRanges(&value.Destinations),
			})
		}
	}
	model.Network.Container = containerNetworks

	// storage
	model.Storage = &models.VCHStorage{
		BaseImageSize: asBytes(&vchConfig.Storage.ScratchSize),
	}

	volumeLocations := make([]string, 0, len(vchConfig.Storage.VolumeLocations))
	for _, value := range vchConfig.Storage.VolumeLocations {
		volumeLocations = append(volumeLocations, value.String())
	}
	model.Storage.VolumeStores = volumeLocations

	imageStores := make([]string, 0, len(vchConfig.Storage.ImageStores))
	for _, value := range vchConfig.Storage.ImageStores {
		imageStores = append(imageStores, value.String())
	}
	model.Storage.ImageStores = imageStores

	// security
	model.Security = &models.VCHSecurity{
		Client:  &models.VCHSecurityClient{},
		Vcenter: &models.VCHSecurityVcenter{},
	}

	if vchConfig.Certificate.HostCertificate != nil {
		model.Security.Server = &models.VCHSecurityServer{
			Certificate: asPemCertificate(vchConfig.Certificate.HostCertificate.Cert),
		}
	}

	clientCertificates := make([]*models.X509Data, 0, len(vchConfig.Certificate.UserCertificates))
	for _, c := range vchConfig.Certificate.UserCertificates {
		clientCertificates = append(clientCertificates, asPemCertificate(c.Cert))
	}
	// TODO: use clientCertificates (should these be unioned with CertificateAuthorities?)

	model.Security.Client.CertificateAuthorities = asPemCertificates(vchConfig.Certificate.CertificateAuthorities)

	// endpoint
	model.Endpoint = &models.VCHEndpoint{
		UseResourcePool: d.UseRP,
		Memory: &models.ValueBytes{
			Value: models.Value{
				Value: int64(d.MemoryMB),
			},
			Units: models.ValueBytesUnitsMiB,
		},
		CPU: &models.VCHEndpointCPU{
			Sockets: int64(d.NumCPUs),
		},
	}

	// registry
	model.Registry = &models.VCHRegistry{
		Insecure:               vchConfig.Registry.InsecureRegistries,
		Whitelist:              vchConfig.Registry.RegistryWhitelist,
		Blacklist:              vchConfig.Registry.RegistryBlacklist,
		CertificateAuthorities: asPemCertificates(vchConfig.Certificate.RegistryCertificateAuthorities),
		ImageFetchProxy:        nil, // TODO: is this information available at runtime?
	}

	// runtime
	model.Runtime = &models.VCHRuntime{}

	installerVer := version.GetBuild()
	upgradeStatus := upgradeStatusMessage(context.Background(), vch, installerVer, vchConfig.Version)
	model.Runtime.UpgradeStatus = upgradeStatus

	powerState, err := vch.PowerState(context.Background())
	if err != nil {
		// TODO
	}
	model.Runtime.PowerState = string(powerState)

	if public := vchConfig.ExecutorConfig.Networks["public"]; public != nil {
		if public_ip := public.Assigned.IP; public_ip != nil {
			var docker_port string
			if !vchConfig.HostCertificate.IsNil() {
				docker_port = fmt.Sprintf("%d", opts.DefaultTLSHTTPPort)
			} else {
				docker_port = fmt.Sprintf("%d", opts.DefaultHTTPPort)
			}

			model.Runtime.DockerHost = fmt.Sprintf("%s:%s", public_ip, docker_port)
			model.Runtime.AdminPortal = fmt.Sprintf("https://%s:2378", public_ip)
		}
	}

	return model
}

func asBytes(value *int64) *models.ValueBytes {
	if value == nil {
		return nil
	}

	if *value == 0 {
		return nil
	}

	return &models.ValueBytes{
		Value: models.Value{
			Value: *value,
			Units: models.ValueBytesUnitsB,
		},
	}
}

func asMHz(value *int64) *models.ValueHertz {
	if value == nil {
		return nil
	}

	if *value == 0 {
		return nil
	}

	return &models.ValueHertz{
		Value: models.Value{
			Value: *value,
			Units: models.ValueHertzUnitsMHz,
		},
	}
}

func asIPAddress(address net.IP) models.IPAddress {
	return models.IPAddress(address.String())
}

func asIPAddresses(addresses *[]net.IP) *[]models.IPAddress {
	m := make([]models.IPAddress, 0, len(*addresses))
	for _, value := range *addresses {
		m = append(m, asIPAddress(value))
	}

	return &m
}

func asIPRange(network *net.IPNet) models.IPRange {
	if network == nil {
		return models.IPRange{}
	}

	return models.IPRange{CIDR: models.CIDR(network.String())}
}

func asIPRanges(networks *[]net.IPNet) *[]models.IPRange {
	m := make([]models.IPRange, 0, len(*networks))
	for _, value := range *networks {
		m = append(m, asIPRange(&value))
	}

	return &m
}

func asNetwork(network *executor.NetworkEndpoint) *models.Network {
	if network == nil {
		return nil
	}

	m := &models.Network{
		PortGroup: &models.ManagedObject{
			ID: network.Network.Common.ID, // TODO: This also doesn't seem to be an ID.
		},
		//Nameservers: asIPAddresses(&network.Network.Nameservers),
	}

	if network.Network.Gateway.IP != nil {
		m.Gateway = &models.Gateway{
			Address:             asIPAddress(network.Network.Gateway.IP),
			RoutingDestinations: *asIPRanges(&network.Network.Destinations),
		}
	}

	return m
}

func asPemCertificates(certificates []byte) []*models.X509Data {
	var buf bytes.Buffer

	m := make([]*models.X509Data, 0)
	for c := &certificates; len(*c) > 0; {
		b, rest := pem.Decode(*c)

		err := pem.Encode(&buf, b)
		if err != nil {
			return nil // TODO: Handle?
		}

		m = append(m, &models.X509Data{
			Pem: buf.String(),
		})

		c = &rest
	}

	return m
}

func asPemCertificate(certificates []byte) *models.X509Data {
	m := asPemCertificates(certificates)

	if len(m) > 1 {
		// Error?
	}

	return m[0]
}
