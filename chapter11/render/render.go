package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"
)

// Model structures (reusing from previous code)
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
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	IPAddress      string   `json:"ip_address,omitempty"`
	SubnetMask     string   `json:"subnet_mask,omitempty"`
	Enabled        bool     `json:"enabled"`
	MTU            int      `json:"mtu,omitempty"`
	Speed          string   `json:"speed"`
	Duplex         string   `json:"duplex"`
	SwitchportMode string   `json:"switchport_mode,omitempty"`
	VLAN           int      `json:"vlan,omitempty"`
	AllowedVLANs   []int    `json:"allowed_vlans,omitempty"`
}

type OSPFArea struct {
	AreaID   string   `json:"area_id"`
	Networks []string `json:"networks"`
}

type BGPNeighbor struct {
	IP          string `json:"ip"`
	RemoteAS    string `json:"remote_as"`
	Description string `json:"description"`
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

type InfrastructureModel struct {
	Metadata Metadata `json:"metadata"`
	Devices  []Device `json:"devices"`
}

// Template function helpers
func getMaskBits(mask string) string {
	masks := map[string]string{
		"255.255.255.252": "30",
		"255.255.255.0":   "24",
		"255.255.0.0":     "16",
		"255.0.0.0":       "8",
	}
	if bits, ok := masks[mask]; ok {
		return bits
	}
	return "32"
}

func join(arr []int, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	result := fmt.Sprintf("%d", arr[0])
	for i := 1; i < len(arr); i++ {
		result += sep + fmt.Sprintf("%d", arr[i])
	}
	return result
}

type ConfigGenerator interface {
	Generate(device *Device, metadata *Metadata) (string, error)
}

type CiscoGenerator struct {
	template *template.Template
}

func NewCiscoGenerator() (*CiscoGenerator, error) {
	funcMap := template.FuncMap{
		"getMaskBits": getMaskBits,
		"join":        join,
	}
	
	tmpl, err := template.New("cisco").Funcs(funcMap).Parse(ciscoTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Cisco template: %w", err)
	}
	
	return &CiscoGenerator{template: tmpl}, nil
}

func (g *CiscoGenerator) Generate(device *Device, metadata *Metadata) (string, error) {
	var buf bytes.Buffer
	
	data := struct {
		*Device
		Metadata *Metadata
	}{
		Device:   device,
		Metadata: metadata,
	}
	
	if err := g.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute Cisco template: %w", err)
	}
	
	return buf.String(), nil
}

type JuniperGenerator struct {
	template *template.Template
}

func NewJuniperGenerator() (*JuniperGenerator, error) {
	funcMap := template.FuncMap{
		"getMaskBits": getMaskBits,
		"join":        join,
	}
	
	tmpl, err := template.New("juniper").Funcs(funcMap).Parse(juniperTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Juniper template: %w", err)
	}
	
	return &JuniperGenerator{template: tmpl}, nil
}

func (g *JuniperGenerator) Generate(device *Device, metadata *Metadata) (string, error) {
	var buf bytes.Buffer
	
	data := struct {
		*Device
		Metadata *Metadata
	}{
		Device:   device,
		Metadata: metadata,
	}
	
	if err := g.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute Juniper template: %w", err)
	}
	
	return buf.String(), nil
}

type GeneratorFactory struct {
	generators map[string]ConfigGenerator
}

func NewGeneratorFactory() (*GeneratorFactory, error) {
	factory := &GeneratorFactory{
		generators: make(map[string]ConfigGenerator),
	}
	
	ciscoGen, err := NewCiscoGenerator()
	if err != nil {
		return nil, err
	}
	factory.generators["cisco"] = ciscoGen
	
	juniperGen, err := NewJuniperGenerator()
	if err != nil {
		return nil, err
	}
	factory.generators["juniper"] = juniperGen
	
	return factory, nil
}

func (f *GeneratorFactory) GetGenerator(vendor string) (ConfigGenerator, error) {
	gen, ok := f.generators[vendor]
	if !ok {
		return nil, fmt.Errorf("no generator found for vendor: %s", vendor)
	}
	return gen, nil
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

func GenerateConfiguration(model *InfrastructureModel, deviceID string) (string, error) {
	var targetDevice *Device
	for i := range model.Devices {
		if model.Devices[i].ID == deviceID {
			targetDevice = &model.Devices[i]
			break
		}
	}
	
	if targetDevice == nil {
		return "", fmt.Errorf("device %s not found", deviceID)
	}
	
	factory, err := NewGeneratorFactory()
	if err != nil {
		return "", fmt.Errorf("failed to create generator factory: %w", err)
	}
	
	generator, err := factory.GetGenerator(targetDevice.Vendor)
	if err != nil {
		return "", err
	}
	
	config, err := generator.Generate(targetDevice, &model.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to generate configuration: %w", err)
	}
	
	return config, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <json-file> [device-id]")
		os.Exit(1)
	}
	
	filename := os.Args[1]
	
	model, err := LoadModel(filename)
	if err != nil {
		fmt.Printf("Error loading model: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Loaded model with %d devices\n\n", len(model.Devices))
	
	if len(os.Args) >= 3 {
		deviceID := os.Args[2]
		config, err := GenerateConfiguration(model, deviceID)
		if err != nil {
			fmt.Printf("Error generating configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(config)
	} else {
		for _, device := range model.Devices {
			fmt.Printf("Generating configuration for %s (%s %s)\n", device.Hostname, device.Vendor, device.Model)
			fmt.Println(strings.Repeat("=", 80))
			
			config, err := GenerateConfiguration(model, device.ID)
			if err != nil {
				fmt.Printf("Error: %v\n\n", err)
				continue
			}
			
			fmt.Println(config)
			fmt.Println()
		}
	}
}