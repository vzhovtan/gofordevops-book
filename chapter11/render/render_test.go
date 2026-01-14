package render

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func createTestDevice(vendor string) *Device {
	device := &Device{
		ID:           "test-device-01",
		Hostname:     "test-router-01",
		DeviceType:   "router",
		Vendor:       vendor,
		Model:        "TestModel",
		ManagementIP: "10.0.1.1",
		Location: Location{
			Datacenter: "dc-test",
			Rack:       "R01",
			Position:   "U10",
		},
		Interfaces: []Interface{
			{
				Name:        "GigabitEthernet0/0/0",
				Description: "Test interface",
				IPAddress:   "192.168.1.1",
				SubnetMask:  "255.255.255.252",
				Enabled:     true,
				MTU:         1500,
				Speed:       "1000",
				Duplex:      "full",
			},
			{
				Name:        "GigabitEthernet0/0/1",
				Description: "Disabled interface",
				IPAddress:   "192.168.1.5",
				SubnetMask:  "255.255.255.252",
				Enabled:     false,
				Speed:       "1000",
				Duplex:      "auto",
			},
		},
		Routing: &Routing{
			Protocols: []RoutingProtocol{
				{
					Protocol:  "ospf",
					ProcessID: "100",
					RouterID:  "1.1.1.1",
					Areas: []OSPFArea{
						{
							AreaID:     "0.0.0.0",
							Networks:   []string{"192.168.1.0/30", "192.168.1.4/30"},
							Interfaces: []string{"GigabitEthernet0/0/0", "GigabitEthernet0/0/1"},
						},
					},
				},
			},
			StaticRoutes: []StaticRoute{
				{
					Destination:            "10.0.0.0/8",
					NextHop:                "192.168.1.2",
					AdministrativeDistance: 1,
				},
			},
		},
		Services: Services{
			NTP: NTPService{
				Enabled: true,
				Servers: []string{"10.0.0.1", "10.0.0.2"},
			},
			SNMP: SNMPService{
				Enabled:   true,
				Community: "public",
				Location:  "Test Lab",
				Contact:   "admin@test.com",
			},
			Syslog: SyslogService{
				Enabled: true,
				Servers: []SyslogServer{
					{
						Host:     "10.0.2.1",
						Port:     514,
						Severity: "informational",
					},
				},
			},
		},
	}

	if vendor == "juniper" {
		device.Interfaces[0].Name = "ge-0/0/0"
		device.Interfaces[1].Name = "ge-0/0/1"
		device.VLANs = []VLAN{
			{
				ID:          100,
				Name:        "test-vlan",
				Description: "Test VLAN",
			},
		}
	}

	return device
}

func createTestMetadata() *Metadata {
	return &Metadata{
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Description: "Test configuration",
		Environment: "test",
	}
}

func TestNewCiscoGenerator(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create Cisco generator: %v", err)
	}

	if gen == nil {
		t.Error("Expected non-nil generator")
	}

	if gen.template == nil {
		t.Error("Expected template to be initialized")
	}
}

func TestNewJuniperGenerator(t *testing.T) {
	gen, err := NewJuniperGenerator()
	if err != nil {
		t.Fatalf("Failed to create Juniper generator: %v", err)
	}

	if gen == nil {
		t.Error("Expected non-nil generator")
	}

	if gen.template == nil {
		t.Error("Expected template to be initialized")
	}
}

func TestCiscoGeneratorBasicConfig(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if config == "" {
		t.Error("Expected non-empty configuration")
	}

	if !strings.Contains(config, "hostname test-router-01") {
		t.Error("Configuration should contain hostname command")
	}
}

func TestCiscoGeneratorInterfaces(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "interface GigabitEthernet0/0/0") {
		t.Error("Configuration should contain interface definition")
	}

	if !strings.Contains(config, "ip address 192.168.1.1 255.255.255.252") {
		t.Error("Configuration should contain IP address")
	}

	if !strings.Contains(config, "description Test interface") {
		t.Error("Configuration should contain interface description")
	}

	if !strings.Contains(config, "mtu 1500") {
		t.Error("Configuration should contain MTU setting")
	}

	if !strings.Contains(config, "no shutdown") {
		t.Error("Configuration should contain no shutdown for enabled interface")
	}

	if !strings.Contains(config, "shutdown") {
		t.Error("Configuration should contain shutdown for disabled interface")
	}
}

func TestCiscoGeneratorNTP(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "ntp server 10.0.0.1") {
		t.Error("Configuration should contain first NTP server")
	}

	if !strings.Contains(config, "ntp server 10.0.0.2") {
		t.Error("Configuration should contain second NTP server")
	}
}

func TestCiscoGeneratorSNMP(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "snmp-server community public") {
		t.Error("Configuration should contain SNMP community")
	}

	if !strings.Contains(config, "snmp-server location Test Lab") {
		t.Error("Configuration should contain SNMP location")
	}

	if !strings.Contains(config, "snmp-server contact admin@test.com") {
		t.Error("Configuration should contain SNMP contact")
	}
}

func TestCiscoGeneratorSyslog(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "logging host 10.0.2.1") {
		t.Error("Configuration should contain syslog host")
	}

	if !strings.Contains(config, "logging trap informational") {
		t.Error("Configuration should contain syslog severity")
	}
}

func TestCiscoGeneratorOSPF(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "router ospf 100") {
		t.Error("Configuration should contain OSPF process")
	}

	if !strings.Contains(config, "router-id 1.1.1.1") {
		t.Error("Configuration should contain router ID")
	}
}

func TestCiscoGeneratorStaticRoutes(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "ip route 10.0.0.0/8 192.168.1.2 1") {
		t.Error("Configuration should contain static route")
	}
}

func TestJuniperGeneratorBasicConfig(t *testing.T) {
	gen, err := NewJuniperGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("juniper")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if config == "" {
		t.Error("Expected non-empty configuration")
	}

	if !strings.Contains(config, "host-name test-router-01") {
		t.Error("Configuration should contain hostname")
	}
}

func TestJuniperGeneratorInterfaces(t *testing.T) {
	gen, err := NewJuniperGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("juniper")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "ge-0/0/0") {
		t.Error("Configuration should contain interface name")
	}

	if !strings.Contains(config, "description \"Test interface\"") {
		t.Error("Configuration should contain interface description")
	}

	if !strings.Contains(config, "address 192.168.1.1/30") {
		t.Error("Configuration should contain IP address with CIDR notation")
	}

	if !strings.Contains(config, "mtu 1500") {
		t.Error("Configuration should contain MTU setting")
	}

	if !strings.Contains(config, "disable") {
		t.Error("Configuration should contain disable for inactive interface")
	}
}

func TestJuniperGeneratorVLANs(t *testing.T) {
	gen, err := NewJuniperGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("juniper")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "vlans {") {
		t.Error("Configuration should contain VLAN section")
	}

	if !strings.Contains(config, "test-vlan") {
		t.Error("Configuration should contain VLAN name")
	}

	if !strings.Contains(config, "vlan-id 100") {
		t.Error("Configuration should contain VLAN ID")
	}
}

func TestJuniperGeneratorNTP(t *testing.T) {
	gen, err := NewJuniperGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("juniper")
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "ntp {") {
		t.Error("Configuration should contain NTP section")
	}

	if !strings.Contains(config, "server 10.0.0.1") {
		t.Error("Configuration should contain NTP server")
	}
}

func TestGetMaskBits(t *testing.T) {
	tests := []struct {
		mask     string
		expected string
	}{
		{"255.255.255.252", "30"},
		{"255.255.255.0", "24"},
		{"255.255.0.0", "16"},
		{"255.0.0.0", "8"},
		{"unknown", "32"},
	}

	for _, tc := range tests {
		result := GetMaskBits(tc.mask)
		if result != tc.expected {
			t.Errorf("getMaskBits(%s) = %s, expected %s", tc.mask, result, tc.expected)
		}
	}
}

func TestJoinFunction(t *testing.T) {
	tests := []struct {
		arr      []int
		sep      string
		expected string
	}{
		{[]int{100, 200, 300}, ",", "100,200,300"},
		{[]int{1}, ",", "1"},
		{[]int{}, ",", ""},
		{[]int{10, 20}, " ", "10 20"},
	}

	for _, tc := range tests {
		result := join(tc.arr, tc.sep)
		if result != tc.expected {
			t.Errorf("join(%v, %s) = %s, expected %s", tc.arr, tc.sep, result, tc.expected)
		}
	}
}

func TestNewGeneratorFactory(t *testing.T) {
	factory, err := NewGeneratorFactory()
	if err != nil {
		t.Fatalf("Failed to create generator factory: %v", err)
	}

	if factory == nil {
		t.Error("Expected non-nil factory")
	}

	if len(factory.generators) != 2 {
		t.Errorf("Expected 2 generators, got %d", len(factory.generators))
	}
}

func TestGeneratorFactoryGetGenerator(t *testing.T) {
	factory, err := NewGeneratorFactory()
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	ciscoGen, err := factory.GetGenerator("cisco")
	if err != nil {
		t.Errorf("Failed to get Cisco generator: %v", err)
	}
	if ciscoGen == nil {
		t.Error("Expected non-nil Cisco generator")
	}

	juniperGen, err := factory.GetGenerator("juniper")
	if err != nil {
		t.Errorf("Failed to get Juniper generator: %v", err)
	}
	if juniperGen == nil {
		t.Error("Expected non-nil Juniper generator")
	}

	_, err = factory.GetGenerator("unknown")
	if err == nil {
		t.Error("Expected error for unknown vendor")
	}
}

func TestGenerateConfigurationIntegration(t *testing.T) {
	model := &InfrastructureModel{
		Metadata: *createTestMetadata(),
		Devices: []Device{
			*createTestDevice("cisco"),
		},
	}

	model.Devices[0].ID = "cisco-test-01"

	config, err := GenerateConfiguration(model, "cisco-test-01")
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if config == "" {
		t.Error("Expected non-empty configuration")
	}

	if !strings.Contains(config, "hostname test-router-01") {
		t.Error("Configuration should contain hostname")
	}
}

func TestGenerateConfigurationDeviceNotFound(t *testing.T) {
	model := &InfrastructureModel{
		Metadata: *createTestMetadata(),
		Devices:  []Device{},
	}

	_, err := GenerateConfiguration(model, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestCiscoGeneratorWithoutRouting(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	device.Routing = nil
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if strings.Contains(config, "router ospf") {
		t.Error("Configuration should not contain OSPF when routing is nil")
	}
}

func TestCiscoGeneratorWithBGP(t *testing.T) {
	gen, err := NewCiscoGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	device := createTestDevice("cisco")
	device.Routing.Protocols = []RoutingProtocol{
		{
			Protocol: "bgp",
			ASNumber: "65001",
			RouterID: "1.1.1.1",
			Neighbors: []BGPNeighbor{
				{
					IP:          "192.168.1.2",
					RemoteAS:    "65002",
					Description: "Test peer",
				},
			},
		},
	}
	metadata := createTestMetadata()

	config, err := gen.Generate(device, metadata)
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "router bgp 65001") {
		t.Error("Configuration should contain BGP configuration")
	}

	if !strings.Contains(config, "neighbor 192.168.1.2 remote-as 65002") {
		t.Error("Configuration should contain BGP neighbor")
	}
}

func TestLoadModelAndGenerate(t *testing.T) {
	model := &InfrastructureModel{
		Metadata: *createTestMetadata(),
		Devices: []Device{
			*createTestDevice("cisco"),
		},
	}

	model.Devices[0].ID = "test-device"

	filename := "test_model_generate.json"
	defer os.Remove(filename)

	data, _ := json.MarshalIndent(model, "", "  ")
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loadedModel, err := LoadModel(filename)
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	config, err := GenerateConfiguration(loadedModel, "test-device")
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	if !strings.Contains(config, "hostname test-router-01") {
		t.Error("Generated configuration should contain hostname")
	}
}
