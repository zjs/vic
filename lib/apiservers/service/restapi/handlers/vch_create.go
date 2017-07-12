package handlers

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"strings"

	"github.com/docker/go-units"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/vmware/govmomi/vim25/types"
	"gopkg.in/urfave/cli.v1"

	"github.com/vmware/vic/cmd/vic-machine/common"
	"github.com/vmware/vic/cmd/vic-machine/create"
	"github.com/vmware/vic/lib/apiservers/service/models"
	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers/util"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
	"github.com/vmware/vic/lib/install/management"
	"github.com/vmware/vic/lib/install/data"
	"github.com/vmware/vic/pkg/version"
)

// VCHCreate is the handler for creating a VCH
type VCHCreate struct {
}

func (h *VCHCreate) Handle(params operations.PostTargetVchParams, principal interface{}) middleware.Responder {
	d := buildData(
		url.URL{Host: params.Target},
		principal.(Credentials).user,
		principal.(Credentials).pass,
		params.Thumbprint,
		nil)

	c, err := buildCreate(d, params)
	if err != nil {
		return operations.NewPostTargetVchDefault(err.Code()).WithPayload(&models.Error{Message: err.Error()})
	}

	task, err := handleCreate(c)
	if err != nil {
		return operations.NewPostTargetVchDefault(err.Code()).WithPayload(&models.Error{Message: err.Error()})
	}

	return operations.NewPostTargetVchCreated().WithPayload(operations.PostTargetVchCreatedBody{Task: task})
}

func buildCreate(d *data.Data, params operations.PostTargetVchParams) (*create.Create, *util.HttpError) {
	c := &create.Create{Data: d}
	// TODO: deduplicate with create.processParams

	if params.Vch.Version != "" && version.String() != string(params.Vch.Version) {
		return nil, util.NewHttpError(400, fmt.Sprintf("Invalid version: %s", params.Vch.Version))
	}

	c.DisplayName = params.Vch.Name

	// TODO: move validation to swagger
	if err := common.CheckUnsupportedChars(c.DisplayName); err != nil {
		return nil, util.NewHttpError(400, fmt.Sprintf("Invalid display name: %s", err))
	}
	if len(c.DisplayName) > create.MaxDisplayNameLen {
		return nil, util.NewHttpError(400, fmt.Sprintf("Invalid display name: length exceeds %d characters", create.MaxDisplayNameLen))
	}

	if params.Vch != nil {
		if params.Vch.Compute != nil {
			if params.Vch.Compute.CPU != nil {
				c.ResourceLimits.VCHCPULimitsMHz = mhzFromValueHertz(params.Vch.Compute.CPU.Limit)
				c.ResourceLimits.VCHCPUReservationsMHz = mhzFromValueHertz(params.Vch.Compute.CPU.Reservation)
				c.ResourceLimits.VCHCPUShares = fromShares(params.Vch.Compute.CPU.Shares)
			}

			if params.Vch.Compute.Memory != nil {
				c.ResourceLimits.VCHMemoryLimitsMB = mbFromValueBytes(params.Vch.Compute.Memory.Limit)
				c.ResourceLimits.VCHMemoryReservationsMB = mbFromValueBytes(params.Vch.Compute.Memory.Reservation)
				c.ResourceLimits.VCHMemoryShares = fromShares(params.Vch.Compute.Memory.Shares)
			}

			c.ComputeResourcePath = fromManagedObject(params.Vch.Compute.Resource)
		}

		if params.Vch.Network != nil {
			c.BridgeIPRange = fromCIDR(&params.Vch.Network.Bridge.IPRange.CIDR)
			c.BridgeNetworkName = fromManagedObject(params.Vch.Network.Bridge.PortGroup)

			if params.Vch.Network.Client != nil {
				c.ClientNetworkName = fromManagedObject(params.Vch.Network.Client.PortGroup)
				c.ClientNetworkGateway = fromGateway(params.Vch.Network.Client.Gateway)
				c.ClientNetworkIP = fromNetworkAddress(params.Vch.Network.Client.Static)
			}

			if params.Vch.Network.Management != nil {
				c.ManagementNetworkName = fromManagedObject(params.Vch.Network.Management.PortGroup)
				c.ManagementNetworkGateway = fromGateway(params.Vch.Network.Management.Gateway)
				c.ManagementNetworkIP = fromNetworkAddress(params.Vch.Network.Management.Static)
			}

			if params.Vch.Network.Public != nil {
				c.PublicNetworkName = fromManagedObject(params.Vch.Network.Public.PortGroup)
				c.PublicNetworkGateway = fromGateway(params.Vch.Network.Public.Gateway)
				c.PublicNetworkIP = fromNetworkAddress(params.Vch.Network.Public.Static)
			}

			//c.ContainerNetworks
		}

		if params.Vch.Storage != nil {
			c.ImageDatastorePath = params.Vch.Storage.ImageStores[0] // TODO: many vs. one mismatch

			vs := common.VolumeStores{VolumeStores: cli.StringSlice(params.Vch.Storage.VolumeStores)}
			volumeLocations, err :=  vs.ProcessVolumeStores()
			if err != nil {
				return nil, util.NewHttpError(400, fmt.Sprintf("Error processing volume stores: %s", err))
			}
			c.VolumeLocations = volumeLocations

			c.ScratchSize = fromValueBytes(params.Vch.Storage.BaseImageSize)
		}

		if params.Vch.Security != nil {
			if params.Vch.Security.Client != nil {
				c.Certs.NoTLS = params.Vch.Security.Client.NoTLS
				c.Certs.NoTLSverify = params.Vch.Security.Client.NoTLSVerify
				c.Certs.ClientCAs = fromPemCertificates(params.Vch.Security.Client.CertificateAuthorities)
				c.ClientCAs = c.Certs.ClientCAs
			}

			if params.Vch.Security.Server != nil {
				if params.Vch.Security.Server.Generate != nil {
					c.Certs.Cname = params.Vch.Security.Server.Generate.Cname
					c.Certs.Org = params.Vch.Security.Server.Generate.Organization
					c.Certs.KeySize = fromValueBits(params.Vch.Security.Server.Generate.Size)

					if err := c.Certs.ProcessCertificates(c.DisplayName, c.Force, 0); err != nil {
						return nil, util.NewHttpError(400, fmt.Sprintf("Error generating certificates: %s", err))
					}
				} else {
					c.Certs.CertPEM = []byte(params.Vch.Security.Server.Certificate.Pem)
					c.Certs.KeyPEM = []byte(params.Vch.Security.Server.PrivateKey.Pem)
				}

				c.CertPEM = c.Certs.CertPEM
				c.KeyPEM = c.Certs.KeyPEM
			}

			if params.Vch.Security.Vcenter != nil {
				if params.Vch.Security.Vcenter.OperationsCredentials != nil {
					opsPassword := string(params.Vch.Security.Vcenter.OperationsCredentials.Password)
					c.OpsCredentials = common.OpsCredentials{
						OpsUser: &params.Vch.Security.Vcenter.OperationsCredentials.User,
						OpsPassword: &opsPassword,
					}
				}
			}
		}

		if params.Vch.Endpoint != nil {
			c.UseRP = params.Vch.Endpoint.UseResourcePool
			c.MemoryMB = *mbFromValueBytes(params.Vch.Endpoint.Memory)
			c.NumCPUs = int(params.Vch.Endpoint.CPU.Sockets)
		}

		if params.Vch.Registry != nil {
			c.InsecureRegistries = params.Vch.Registry.Insecure
			c.WhitelistRegistries = params.Vch.Registry.Whitelist

			//params.Vch.Registry.Blacklist

			c.RegistryCAs = fromPemCertificates(params.Vch.Registry.CertificateAuthorities)

			if params.Vch.Registry.ImageFetchProxy != nil {
				c.Proxies = common.Proxies{
					HTTPProxy: &params.Vch.Registry.ImageFetchProxy.HTTP,
					HTTPSProxy: &params.Vch.Registry.ImageFetchProxy.HTTPS,
				}
				_, _, err := c.Proxies.ProcessProxies()
				if err != nil {
					return nil, util.NewHttpError(400, fmt.Sprintf("Error processing proxies: %s", err))
				}
			}
		}
	}

	return c, nil
}

func handleCreate(c *create.Create) (*strfmt.URI, *util.HttpError) {
	validator, err := validateTarget(c.Data)
	if err != nil {
		return nil, util.NewHttpError(400, err.Error())
	}

	vchConfig, err := validator.Validate(validator.Context, c.Data)
	vConfig := validator.AddDeprecatedFields(validator.Context, vchConfig, c.Data)

	// TODO: make this configurable
	images := common.Images{}
	vConfig.ImageFiles, err = images.CheckImagesFiles(true)
	vConfig.ApplianceISO = path.Base(images.ApplianceISO)
	vConfig.BootstrapISO = path.Base(images.BootstrapISO)

	// TODO: this doesn't seem right
	vConfig.HTTPProxy = c.HTTPProxy
	vConfig.HTTPSProxy = c.HTTPSProxy

	executor := management.NewDispatcher(validator.Context, validator.Session, nil, false)
	err = executor.CreateVCH(vchConfig, vConfig)
	if err != nil {
		return nil, util.NewHttpError(500, fmt.Sprintf("Failed to create VCH: %s", err))
	}

	return nil, nil
}

func fromIPRanges(m *[]models.IPRange) *[]string {
	s := make([]string, 0, len(*m))
	for _, d := range *m {
		s = append(s, string(d.CIDR))
	}

	return &s
}

func fromNetworkAddress(m *models.NetworkAddress) string {
	if m == nil {
		return ""
	}

	if m.IP != "" {
		return string(m.IP)
	}

	return string(m.Hostname)
}

func fromManagedObject(m *models.ManagedObject) string {
	if m.ID != "" {
		return m.ID
	}

	return m.Name
}

func fromCIDR(m *models.CIDR) string {
	return string(*m)
}

func fromGateway(m *models.Gateway) string {
	if m == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", // TODO: what if RoutingDestinations is empty?
		strings.Join(*fromIPRanges(&m.RoutingDestinations), ","),
		m.Address,
	)
}

func fromValueBytes(m *models.ValueBytes) string {
	return fmt.Sprintf("%d%s", m.Value.Value, m.Units)
}

func mbFromValueBytes(m *models.ValueBytes) *int {
	v := float64(m.Value.Value)

	var mbs float64
	switch m.Units {
	case models.ValueBytesUnitsB:
		mbs = v / float64(units.MiB)
	case models.ValueBytesUnitsKiB:
		mbs = v / (float64(units.MiB) / float64(units.KiB))
	case models.ValueBytesUnitsMiB:
		mbs = v
	case models.ValueBytesUnitsGiB:
		mbs = v * (float64(units.GiB) / float64(units.MiB))
	case models.ValueBytesUnitsTiB:
		mbs = v * (float64(units.TiB) / float64(units.MiB))
	case models.ValueBytesUnitsPiB:
		mbs = v * (float64(units.PiB) / float64(units.MiB))
	}

	i := int(math.Ceil(mbs))

	return &i
}

func mhzFromValueHertz(m *models.ValueHertz) *int {
	v := float64(m.Value.Value)

	var mhzs float64
	switch m.Units {
	case models.ValueHertzUnitsHz:
		mhzs = v / float64(units.MB)
	case models.ValueHertzUnitsKHz:
		mhzs = v / (float64(units.MB) / float64(units.KB))
	case models.ValueHertzUnitsMHz:
		mhzs = v
	case models.ValueHertzUnitsGHz:
		mhzs = v * (float64(units.GB) / float64(units.MB))
	}

	i := int(math.Ceil(mhzs))

	return  &i
}

func fromShares(m *models.Shares) *types.SharesInfo {

	var level types.SharesLevel
	switch types.SharesLevel(m.Level) {
	case types.SharesLevelLow:
		level = types.SharesLevelLow
	case types.SharesLevelNormal:
		level = types.SharesLevelNormal
	case types.SharesLevelHigh:
		level = types.SharesLevelHigh
	default:
		level = types.SharesLevelCustom
	}

	return &types.SharesInfo{
		Level: level,
		Shares: int32(m.Number),
	}
}

func fromValueBits(m *models.ValueBits) int {
	return int(m.Value.Value)
}

func fromPemCertificates(m []*models.X509Data) []byte {
	var b []byte

	for _, ca := range m {
		c := []byte(ca.Pem)
		b = append(b, c...)
	}

	return b
}
