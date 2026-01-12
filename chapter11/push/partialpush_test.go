package push

import (
	"errors"
	"strings"
	"testing"
	"time"
	"model"
	"render"
)

func createJuniperTestDevice() *model.Device {
	return &model.Device{
		ID:           "juniper-test-01",
		Hostname:     "test-juniper-switch",
		DeviceType:   "switch",
		Vendor:       "juniper",
		Model:        "EX4300",
		ManagementIP: "10.0.1.20",
		Location: Location{
			Datacenter: "dc-test",
			Rack:       "R01",
			Position:   "U10",
		},
		Interfaces: []Interface{
			{
				Name:        "ge-0/0/0",
				Description: "Test interface",
				IPAddress:   "192.168.1.1",
				SubnetMask:  "255.255.255.252",
				Enabled:     true,
				MTU:         1500,
				Speed:       "1000",
				Duplex:      "full",
			},
			{
				Name:        "ge-0/0/1",
				Description: "Disabled interface",
				Enabled:     false,
				Speed:       "1000",
				Duplex:      "auto",
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
		},
		VLANs: []VLAN{
			{
				ID:          100,
				Name:        "test-vlan",
				Description: "Test VLAN",
			},
			{
				ID:          200,
				Name:        "mgmt-vlan",
				Description: "Management VLAN",
			},
		},
		Routing: &Routing{
			StaticRoutes: []StaticRoute{
				{
					Destination: "10.0.0.0/8",
					NextHop:     "192.168.1.2",
				},
			},
		},
	}
}

func TestNewPerElementStrategy(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)

	if strategy == nil {
		t.Fatal("Expected non-nil strategy")
	}

	if strategy.sshConfig == nil {
		t.Error("Expected SSH config to be initialized")
	}

	if strategy.sshConfig.User != "admin" {
		t.Errorf("Expected user 'admin', got %s", strategy.sshConfig.User)
	}

	if strategy.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", strategy.timeout)
	}
}

func TestNewPerElementDeployer(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	deployer := NewPerElementDeployer(strategy)

	if deployer == nil {
		t.Fatal("Expected non-nil deployer")
	}

	if deployer.strategy == nil {
		t.Error("Expected strategy to be set")
	}
}

func TestParseConfigToElementsInterfaces(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	foundDescription := false
	foundIPAddress := false
	foundMTU := false
	foundDisable := false

	for _, element := range elements {
		if element.Type == "interface" {
			if strings.Contains(element.Value, "description \"Test interface\"") {
				foundDescription = true
			}
			if strings.Contains(element.Value, "address 192.168.1.1/30") {
				foundIPAddress = true
			}
			if strings.Contains(element.Value, "mtu 1500") {
				foundMTU = true
			}
			if element.Value == "disable" && strings.Contains(element.Path, "ge-0/0/1") {
				foundDisable = true
			}
		}
	}

	if !foundDescription {
		t.Error("Expected to find interface description element")
	}
	if !foundIPAddress {
		t.Error("Expected to find IP address element with CIDR notation")
	}
	if !foundMTU {
		t.Error("Expected to find MTU element")
	}
	if !foundDisable {
		t.Error("Expected to find disable element for disabled interface")
	}
}

func TestParseConfigToElementsNTP(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	ntpCount := 0
	for _, element := range elements {
		if element.Type == "service" && strings.Contains(element.Path, "ntp") {
			ntpCount++
		}
	}

	if ntpCount != 2 {
		t.Errorf("Expected 2 NTP server elements, got %d", ntpCount)
	}
}

func TestParseConfigToElementsSNMP(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	foundCommunity := false
	foundLocation := false
	foundContact := false

	for _, element := range elements {
		if element.Type == "service" && strings.Contains(element.Path, "snmp") {
			if strings.Contains(element.Value, "authorization read-only") {
				foundCommunity = true
			}
			if strings.Contains(element.Value, "location \"Test Lab\"") {
				foundLocation = true
			}
			if strings.Contains(element.Value, "contact \"admin@test.com\"") {
				foundContact = true
			}
		}
	}

	if !foundCommunity {
		t.Error("Expected to find SNMP community element")
	}
	if !foundLocation {
		t.Error("Expected to find SNMP location element")
	}
	if !foundContact {
		t.Error("Expected to find SNMP contact element")
	}
}

func TestParseConfigToElementsVLANs(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	vlanElements := 0
	foundVLAN100 := false
	foundVLAN200 := false

	for _, element := range elements {
		if element.Type == "vlan" {
			vlanElements++
			if strings.Contains(element.Path, "test-vlan") && strings.Contains(element.Value, "vlan-id 100") {
				foundVLAN100 = true
			}
			if strings.Contains(element.Path, "mgmt-vlan") && strings.Contains(element.Value, "vlan-id 200") {
				foundVLAN200 = true
			}
		}
	}

	if vlanElements < 2 {
		t.Errorf("Expected at least 2 VLAN elements, got %d", vlanElements)
	}

	if !foundVLAN100 {
		t.Error("Expected to find VLAN 100 configuration")
	}

	if !foundVLAN200 {
		t.Error("Expected to find VLAN 200 configuration")
	}
}

func TestParseConfigToElementsStaticRoutes(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	foundRoute := false
	for _, element := range elements {
		if element.Type == "routing" {
			if strings.Contains(element.Value, "route 10.0.0.0/8 next-hop 192.168.1.2") {
				foundRoute = true
			}
		}
	}

	if !foundRoute {
		t.Error("Expected to find static route element")
	}
}

func TestParseConfigToElementsNoRouting(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()
	device.Routing = nil

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	for _, element := range elements {
		if element.Type == "routing" {
			t.Error("Should not have routing elements when Routing is nil")
		}
	}
}

func TestConfigElementStructure(t *testing.T) {
	element := ConfigElement{
		Type:        "interface",
		Path:        "interfaces ge-0/0/0",
		Operation:   "set",
		Value:       "description \"Test\"",
		Description: "Set interface description",
	}

	if element.Type != "interface" {
		t.Errorf("Expected type 'interface', got %s", element.Type)
	}

	if element.Operation != "set" {
		t.Errorf("Expected operation 'set', got %s", element.Operation)
	}

	if element.Path == "" {
		t.Error("Expected non-empty path")
	}

	if element.Description == "" {
		t.Error("Expected non-empty description")
	}
}

func TestBuildJuniperCommands(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)

	elements := []ConfigElement{
		{
			Type:      "interface",
			Path:      "interfaces ge-0/0/0",
			Operation: "set",
			Value:     "description \"Test\"",
		},
		{
			Type:      "interface",
			Path:      "interfaces ge-0/0/0",
			Operation: "set",
			Value:     "mtu 9000",
		},
		{
			Type:      "interface",
			Path:      "interfaces ge-0/0/1",
			Operation: "delete",
			Value:     "",
		},
	}

	commands := strategy.buildJuniperCommands(elements)

	if len(commands) < 3 {
		t.Errorf("Expected at least 3 commands, got %d", len(commands))
	}

	if commands[0] != "configure" {
		t.Errorf("Expected first command to be 'configure', got %s", commands[0])
	}

	lastCommand := commands[len(commands)-1]
	if lastCommand != "commit and-quit" {
		t.Errorf("Expected last command to be 'commit and-quit', got %s", lastCommand)
	}

	foundSet := false
	foundDelete := false

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, "set") {
			foundSet = true
		}
		if strings.HasPrefix(cmd, "delete") {
			foundDelete = true
		}
	}

	if !foundSet {
		t.Error("Expected to find 'set' command")
	}

	if !foundDelete {
		t.Error("Expected to find 'delete' command")
	}
}

func TestElementUpdateResultStructure(t *testing.T) {
	element := ConfigElement{
		Type:      "interface",
		Path:      "interfaces ge-0/0/0",
		Operation: "set",
		Value:     "mtu 1500",
	}

	result := ElementUpdateResult{
		Element:   element,
		Success:   true,
		Error:     nil,
		Duration:  2 * time.Second,
		Timestamp: time.Now(),
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Error != nil {
		t.Error("Expected no error")
	}

	if result.Duration != 2*time.Second {
		t.Errorf("Expected duration 2s, got %v", result.Duration)
	}

	if result.Element.Type != "interface" {
		t.Errorf("Expected element type 'interface', got %s", result.Element.Type)
	}
}

func TestElementUpdateResultWithError(t *testing.T) {
	element := ConfigElement{
		Type:      "interface",
		Path:      "interfaces ge-0/0/0",
		Operation: "set",
		Value:     "invalid config",
	}

	testError := errors.New("configuration error")
	result := ElementUpdateResult{
		Element:   element,
		Success:   false,
		Error:     testError,
		Duration:  1 * time.Second,
		Timestamp: time.Now(),
	}

	if result.Success {
		t.Error("Expected success to be false")
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}

	if result.Error != testError {
		t.Errorf("Expected error %v, got %v", testError, result.Error)
	}
}

func TestParseConfigToElementsEmptyDevice(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := &Device{
		ID:       "empty-device",
		Hostname: "empty",
		Vendor:   "juniper",
	}

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(elements) != 0 {
		t.Errorf("Expected 0 elements for empty device, got %d", len(elements))
	}
}

func TestParseConfigMultipleInterfaceProperties(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	interfaceElements := 0
	for _, element := range elements {
		if element.Type == "interface" && strings.Contains(element.Path, "ge-0/0/0") {
			interfaceElements++
		}
	}

	if interfaceElements < 3 {
		t.Errorf("Expected at least 3 elements for ge-0/0/0 (description, IP, MTU), got %d", interfaceElements)
	}
}

func TestConfigElementOperations(t *testing.T) {
	tests := []struct {
		operation string
		valid     bool
	}{
		{"set", true},
		{"delete", true},
		{"edit", true},
		{"invalid", false},
	}

	for _, tc := range tests {
		element := ConfigElement{
			Operation: tc.operation,
		}

		if tc.valid {
			if element.Operation != tc.operation {
				t.Errorf("Expected operation %s, got %s", tc.operation, element.Operation)
			}
		}
	}
}

func TestDeployToNonJuniperDevice(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createTestDeviceForDeployment("cisco")

	err := strategy.Deploy(device, "config")

	if err == nil {
		t.Error("Expected error when deploying to non-Juniper device")
	}

	if !strings.Contains(err.Error(), "only supported for Juniper") {
		t.Errorf("Expected vendor error message, got: %v", err)
	}
}

func TestVLANElementsWithDescriptions(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	vlanDescriptions := 0
	for _, element := range elements {
		if element.Type == "vlan" && strings.Contains(element.Value, "description") {
			vlanDescriptions++
		}
	}

	if vlanDescriptions != 2 {
		t.Errorf("Expected 2 VLAN description elements, got %d", vlanDescriptions)
	}
}

func TestInterfaceDisableElement(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	foundDisable := false
	for _, element := range elements {
		if element.Type == "interface" &&
			strings.Contains(element.Path, "ge-0/0/1") &&
			element.Value == "disable" {
			foundDisable = true
			if element.Operation != "set" {
				t.Errorf("Expected 'set' operation for disable, got %s", element.Operation)
			}
		}
	}

	if !foundDisable {
		t.Error("Expected to find disable element for ge-0/0/1")
	}
}

func TestEnabledInterfaceNoDisable(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	for _, element := range elements {
		if element.Type == "interface" &&
			strings.Contains(element.Path, "ge-0/0/0") &&
			element.Value == "disable" {
			t.Error("Should not have disable element for enabled interface ge-0/0/0")
		}
	}
}

func TestMultipleNTPServers(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	ntpServers := make(map[string]bool)
	for _, element := range elements {
		if element.Type == "service" && strings.Contains(element.Path, "ntp") {
			if strings.Contains(element.Value, "10.0.0.1") {
				ntpServers["10.0.0.1"] = true
			}
			if strings.Contains(element.Value, "10.0.0.2") {
				ntpServers["10.0.0.2"] = true
			}
		}
	}

	if len(ntpServers) != 2 {
		t.Errorf("Expected 2 unique NTP servers, got %d", len(ntpServers))
	}
}

func TestParseConfigElementPaths(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	for _, element := range elements {
		if element.Path == "" {
			t.Errorf("Element should have non-empty path: %+v", element)
		}

		if !strings.Contains(element.Path, " ") && element.Type != "service" {
			t.Errorf("Expected hierarchical path with spaces, got: %s", element.Path)
		}
	}
}

func TestIPAddressCIDRConversion(t *testing.T) {
	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	device := createJuniperTestDevice()

	elements, err := strategy.parseConfigToElements("", device)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	foundCIDR := false
	for _, element := range elements {
		if element.Type == "interface" && strings.Contains(element.Value, "address") {
			if strings.Contains(element.Value, "/30") {
				foundCIDR = true
			}
			if strings.Contains(element.Value, "255.255.255.252") {
				t.Error("Should use CIDR notation, not subnet mask")
			}
		}
	}

	if !foundCIDR {
		t.Error("Expected to find IP address with CIDR notation (/30)")
	}
}