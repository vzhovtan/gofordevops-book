package model

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Metadata struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	Environment string    `json:"environment"`
}

type Location struct {
	Datacenter string `json:"datacenter"`
	Rack       string `json:"rack"`
	Position   string `json:"position"`
}

type Interface struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	IPAddress      string `json:"ip_address,omitempty"`
	SubnetMask     string `json:"subnet_mask,omitempty"`
	Enabled        bool   `json:"enabled"`
	MTU            int    `json:"mtu,omitempty"`
	Speed          string `json:"speed"`
	Duplex         string `json:"duplex"`
	SwitchportMode string `json:"switchport_mode,omitempty"`
	VLAN           int    `json:"vlan,omitempty"`
	AllowedVLANs   []int  `json:"allowed_vlans,omitempty"`
}

type OSPFArea struct {
	AreaID   string   `json:"area_id"`
	Networks []string `json:"networks"`
}

type OSPFProtocol struct {
	Protocol  string     `json:"protocol"`
	ProcessID string     `json:"process_id"`
	RouterID  string     `json:"router_id"`
	Areas     []OSPFArea `json:"areas"`
}

type BGPNeighbor struct {
	IP          string `json:"ip"`
	RemoteAS    string `json:"remote_as"`
	Description string `json:"description"`
}

type BGPProtocol struct {
	Protocol  string        `json:"protocol"`
	ASNumber  string        `json:"as_number"`
	RouterID  string        `json:"router_id"`
	Neighbors []BGPNeighbor `json:"neighbors"`
}

type RoutingProtocol struct {
	Protocol  string        `json:"protocol"`
	ProcessID string        `json:"process_id,omitempty"`
	RouterID  string        `json:"router_id,omitempty"`
	ASNumber  string        `json:"as_number,omitempty"`
	Areas     []OSPFArea    `json:"areas,omitempty"`
	Neighbors []BGPNeighbor `json:"neighbors,omitempty"`
}

type StaticRoute struct {
	Destination            string `json:"destination"`
	NextHop                string `json:"next_hop"`
	AdministrativeDistance int    `json:"administrative_distance"`
}

type Routing struct {
	Protocols    []RoutingProtocol `json:"protocols"`
	StaticRoutes []StaticRoute     `json:"static_routes"`
}

type NTPService struct {
	Enabled bool     `json:"enabled"`
	Servers []string `json:"servers"`
}

type SNMPService struct {
	Enabled   bool   `json:"enabled"`
	Community string `json:"community"`
	Location  string `json:"location"`
	Contact   string `json:"contact"`
}

type SyslogServer struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Severity string `json:"severity"`
}

type SyslogService struct {
	Enabled bool           `json:"enabled"`
	Servers []SyslogServer `json:"servers"`
}

type Services struct {
	NTP    NTPService    `json:"ntp"`
	SNMP   SNMPService   `json:"snmp"`
	Syslog SyslogService `json:"syslog"`
}

type VLAN struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Device struct {
	ID           string      `json:"id"`
	Hostname     string      `json:"hostname"`
	DeviceType   string      `json:"device_type"`
	Vendor       string      `json:"vendor"`
	Model        string      `json:"model"`
	ManagementIP string      `json:"management_ip"`
	Location     Location    `json:"location"`
	Interfaces   []Interface `json:"interfaces"`
	Routing      *Routing    `json:"routing,omitempty"`
	Services     Services    `json:"services"`
	VLANs        []VLAN      `json:"vlans,omitempty"`
}

type ACLEntry struct {
	Sequence        int    `json:"sequence"`
	Action          string `json:"action"`
	Protocol        string `json:"protocol"`
	Source          string `json:"source"`
	Destination     string `json:"destination"`
	DestinationPort int    `json:"destination_port,omitempty"`
}

type AccessList struct {
	Name    string     `json:"name"`
	Entries []ACLEntry `json:"entries"`
}

type Security struct {
	AccessLists []AccessList `json:"access_lists"`
}

type InfrastructureModel struct {
	Metadata Metadata `json:"metadata"`
	Devices  []Device `json:"devices"`
	Security Security `json:"security"`
}

func LoadModel(filename string) (*InfrastructureModel, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var model InfrastructureModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &model, nil
}

func SaveModel(filename string, model *InfrastructureModel) error {
	model.Metadata.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func UpdateDeviceInterface(model *InfrastructureModel, deviceID, interfaceName string, updates map[string]interface{}) error {
	for i := range model.Devices {
		if model.Devices[i].ID == deviceID {
			for j := range model.Devices[i].Interfaces {
				if model.Devices[i].Interfaces[j].Name == interfaceName {
					iface := &model.Devices[i].Interfaces[j]

					if desc, ok := updates["description"].(string); ok {
						iface.Description = desc
					}
					if enabled, ok := updates["enabled"].(bool); ok {
						iface.Enabled = enabled
					}
					if ip, ok := updates["ip_address"].(string); ok {
						iface.IPAddress = ip
					}
					if mask, ok := updates["subnet_mask"].(string); ok {
						iface.SubnetMask = mask
					}
					if mtu, ok := updates["mtu"].(int); ok {
						iface.MTU = mtu
					}

					return nil
				}
			}
			return fmt.Errorf("interface %s not found on device %s", interfaceName, deviceID)
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func AddStaticRoute(model *InfrastructureModel, deviceID string, route StaticRoute) error {
	for i := range model.Devices {
		if model.Devices[i].ID == deviceID {
			if model.Devices[i].Routing == nil {
				model.Devices[i].Routing = &Routing{}
			}
			model.Devices[i].Routing.StaticRoutes = append(
				model.Devices[i].Routing.StaticRoutes,
				route,
			)
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}

func GetDeviceByID(model *InfrastructureModel, deviceID string) (*Device, error) {
	for i := range model.Devices {
		if model.Devices[i].ID == deviceID {
			return &model.Devices[i], nil
		}
	}
	return nil, fmt.Errorf("device %s not found", deviceID)
}

func ListDevicesByVendor(model *InfrastructureModel, vendor string) []Device {
	var devices []Device
	for _, device := range model.Devices {
		if device.Vendor == vendor {
			devices = append(devices, device)
		}
	}
	return devices
}

func UpdateDeviceManagementIP(model *InfrastructureModel, deviceID, newIP string) error {
	for i := range model.Devices {
		if model.Devices[i].ID == deviceID {
			model.Devices[i].ManagementIP = newIP
			return nil
		}
	}
	return fmt.Errorf("device %s not found", deviceID)
}
