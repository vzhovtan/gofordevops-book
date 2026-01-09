package model

import (
	"os"
	"testing"
	"time"
)

func createTestModel() *InfrastructureModel {
	return &InfrastructureModel{
		Metadata: Metadata{
			Version:     "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Description: "Test infrastructure model",
			Environment: "test",
		},
		Devices: []Device{
			{
				ID:           "test-rtr-01",
				Hostname:     "test-router-01.example.com",
				DeviceType:   "router",
				Vendor:       "cisco",
				Model:        "ASR9000",
				ManagementIP: "10.0.1.10",
				Location: Location{
					Datacenter: "dc-test-01",
					Rack:       "R10",
					Position:   "U20",
				},
				Interfaces: []Interface{
					{
						Name:        "GigabitEthernet0/0/0",
						Description: "Test uplink",
						IPAddress:   "192.168.1.1",
						SubnetMask:  "255.255.255.252",
						Enabled:     true,
						MTU:         1500,
						Speed:       "1000",
						Duplex:      "full",
					},
					{
						Name:        "GigabitEthernet0/0/1",
						Description: "Test connection",
						IPAddress:   "192.168.1.5",
						SubnetMask:  "255.255.255.252",
						Enabled:     true,
						Speed:       "1000",
						Duplex:      "full",
					},
				},
				Routing: &Routing{
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
						Servers: []string{"10.0.0.1"},
					},
				},
			},
			{
				ID:           "test-sw-01",
				Hostname:     "test-switch-01.example.com",
				DeviceType:   "switch",
				Vendor:       "juniper",
				Model:        "EX4300",
				ManagementIP: "10.0.1.20",
				Location: Location{
					Datacenter: "dc-test-01",
					Rack:       "R11",
					Position:   "U10",
				},
				Interfaces: []Interface{
					{
						Name:           "ge-0/0/0",
						Description:    "Access port",
						Enabled:        true,
						SwitchportMode: "access",
						VLAN:           100,
						Speed:          "1000",
						Duplex:         "full",
					},
				},
				Services: Services{
					NTP: NTPService{
						Enabled: true,
						Servers: []string{"10.0.0.1"},
					},
				},
				VLANs: []VLAN{
					{
						ID:          100,
						Name:        "test-vlan",
						Description: "Test VLAN",
					},
				},
			},
		},
		Security: Security{
			AccessLists: []AccessList{
				{
					Name: "TEST-ACL",
					Entries: []ACLEntry{
						{
							Sequence:    10,
							Action:      "permit",
							Protocol:    "tcp",
							Source:      "10.0.0.0/8",
							Destination: "any",
						},
					},
				},
			},
		},
	}
}

func TestLoadModel(t *testing.T) {
	model := createTestModel()
	filename := "test_infrastructure.json"
	
	defer os.Remove(filename)
	
	err := SaveModel(filename, model)
	if err != nil {
		t.Fatalf("Failed to save test model: %v", err)
	}
	
	loadedModel, err := LoadModel(filename)
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	
	if loadedModel.Metadata.Version != model.Metadata.Version {
		t.Errorf("Expected version %s, got %s", model.Metadata.Version, loadedModel.Metadata.Version)
	}
	
	if len(loadedModel.Devices) != len(model.Devices) {
		t.Errorf("Expected %d devices, got %d", len(model.Devices), len(loadedModel.Devices))
	}
	
	if loadedModel.Devices[0].Hostname != model.Devices[0].Hostname {
		t.Errorf("Expected hostname %s, got %s", model.Devices[0].Hostname, loadedModel.Devices[0].Hostname)
	}
}

func TestLoadModelFileNotFound(t *testing.T) {
	_, err := LoadModel("nonexistent_file.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent file, got nil")
	}
}

func TestSaveModel(t *testing.T) {
	model := createTestModel()
	filename := "test_save_infrastructure.json"
	
	defer os.Remove(filename)
	
	originalUpdateTime := model.Metadata.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	
	err := SaveModel(filename, model)
	if err != nil {
		t.Fatalf("Failed to save model: %v", err)
	}
	
	if !model.Metadata.UpdatedAt.After(originalUpdateTime) {
		t.Error("Expected UpdatedAt timestamp to be updated")
	}
	
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to exist after save")
	}
}

func TestGetDeviceByID(t *testing.T) {
	model := createTestModel()
	
	device, err := GetDeviceByID(model, "test-rtr-01")
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}
	
	if device.ID != "test-rtr-01" {
		t.Errorf("Expected device ID test-rtr-01, got %s", device.ID)
	}
	
	if device.Vendor != "cisco" {
		t.Errorf("Expected vendor cisco, got %s", device.Vendor)
	}
}

func TestGetDeviceByIDNotFound(t *testing.T) {
	model := createTestModel()
	
	_, err := GetDeviceByID(model, "nonexistent-device")
	if err == nil {
		t.Error("Expected error for nonexistent device, got nil")
	}
}

func TestListDevicesByVendor(t *testing.T) {
	model := createTestModel()
	
	ciscoDevices := ListDevicesByVendor(model, "cisco")
	if len(ciscoDevices) != 1 {
		t.Errorf("Expected 1 Cisco device, got %d", len(ciscoDevices))
	}
	
	if ciscoDevices[0].ID != "test-rtr-01" {
		t.Errorf("Expected device ID test-rtr-01, got %s", ciscoDevices[0].ID)
	}
	
	juniperDevices := ListDevicesByVendor(model, "juniper")
	if len(juniperDevices) != 1 {
		t.Errorf("Expected 1 Juniper device, got %d", len(juniperDevices))
	}
	
	aristaDevices := ListDevicesByVendor(model, "arista")
	if len(aristaDevices) != 0 {
		t.Errorf("Expected 0 Arista devices, got %d", len(aristaDevices))
	}
}

func TestUpdateDeviceInterface(t *testing.T) {
	model := createTestModel()
	
	updates := map[string]interface{}{
		"description": "Updated test uplink",
		"mtu":         9000,
		"enabled":     false,
	}
	
	err := UpdateDeviceInterface(model, "test-rtr-01", "GigabitEthernet0/0/0", updates)
	if err != nil {
		t.Fatalf("Failed to update interface: %v", err)
	}
	
	device, _ := GetDeviceByID(model, "test-rtr-01")
	iface := device.Interfaces[0]
	
	if iface.Description != "Updated test uplink" {
		t.Errorf("Expected description 'Updated test uplink', got %s", iface.Description)
	}
	
	if iface.MTU != 9000 {
		t.Errorf("Expected MTU 9000, got %d", iface.MTU)
	}
	
	if iface.Enabled != false {
		t.Error("Expected interface to be disabled")
	}
}

func TestUpdateDeviceInterfaceNotFound(t *testing.T) {
	model := createTestModel()
	
	updates := map[string]interface{}{
		"description": "Test",
	}
	
	err := UpdateDeviceInterface(model, "test-rtr-01", "NonexistentInterface", updates)
	if err == nil {
		t.Error("Expected error for nonexistent interface, got nil")
	}
	
	err = UpdateDeviceInterface(model, "nonexistent-device", "GigabitEthernet0/0/0", updates)
	if err == nil {
		t.Error("Expected error for nonexistent device, got nil")
	}
}

func TestAddStaticRoute(t *testing.T) {
	model := createTestModel()
	
	newRoute := StaticRoute{
		Destination:            "172.16.0.0/16",
		NextHop:                "192.168.1.2",
		AdministrativeDistance: 5,
	}
	
	err := AddStaticRoute(model, "test-rtr-01", newRoute)
	if err != nil {
		t.Fatalf("Failed to add static route: %v", err)
	}
	
	device, _ := GetDeviceByID(model, "test-rtr-01")
	
	if len(device.Routing.StaticRoutes) != 2 {
		t.Errorf("Expected 2 static routes, got %d", len(device.Routing.StaticRoutes))
	}
	
	lastRoute := device.Routing.StaticRoutes[len(device.Routing.StaticRoutes)-1]
	if lastRoute.Destination != "172.16.0.0/16" {
		t.Errorf("Expected destination 172.16.0.0/16, got %s", lastRoute.Destination)
	}
	
	if lastRoute.AdministrativeDistance != 5 {
		t.Errorf("Expected administrative distance 5, got %d", lastRoute.AdministrativeDistance)
	}
}

func TestAddStaticRouteToDeviceWithoutRouting(t *testing.T) {
	model := createTestModel()
	
	model.Devices[1].Routing = nil
	
	newRoute := StaticRoute{
		Destination:            "10.10.0.0/16",
		NextHop:                "10.0.1.1",
		AdministrativeDistance: 1,
	}
	
	err := AddStaticRoute(model, "test-sw-01", newRoute)
	if err != nil {
		t.Fatalf("Failed to add static route: %v", err)
	}
	
	device, _ := GetDeviceByID(model, "test-sw-01")
	
	if device.Routing == nil {
		t.Fatal("Expected Routing to be initialized")
	}
	
	if len(device.Routing.StaticRoutes) != 1 {
		t.Errorf("Expected 1 static route, got %d", len(device.Routing.StaticRoutes))
	}
}

func TestAddStaticRouteDeviceNotFound(t *testing.T) {
	model := createTestModel()
	
	newRoute := StaticRoute{
		Destination: "10.0.0.0/8",
		NextHop:     "192.168.1.1",
	}
	
	err := AddStaticRoute(model, "nonexistent-device", newRoute)
	if err == nil {
		t.Error("Expected error for nonexistent device, got nil")
	}
}

func TestUpdateDeviceManagementIP(t *testing.T) {
	model := createTestModel()
	
	err := UpdateDeviceManagementIP(model, "test-rtr-01", "10.0.2.10")
	if err != nil {
		t.Fatalf("Failed to update management IP: %v", err)
	}
	
	device, _ := GetDeviceByID(model, "test-rtr-01")
	
	if device.ManagementIP != "10.0.2.10" {
		t.Errorf("Expected management IP 10.0.2.10, got %s", device.ManagementIP)
	}
}

func TestUpdateDeviceManagementIPNotFound(t *testing.T) {
	model := createTestModel()
	
	err := UpdateDeviceManagementIP(model, "nonexistent-device", "10.0.2.10")
	if err == nil {
		t.Error("Expected error for nonexistent device, got nil")
	}
}

func TestRoundTripSaveAndLoad(t *testing.T) {
	originalModel := createTestModel()
	filename := "test_roundtrip.json"
	
	defer os.Remove(filename)
	
	updates := map[string]interface{}{
		"description": "Modified interface",
		"mtu":         8000,
	}
	err := UpdateDeviceInterface(originalModel, "test-rtr-01", "GigabitEthernet0/0/0", updates)
	if err != nil {
		t.Fatalf("Failed to update interface: %v", err)
	}
	
	err = SaveModel(filename, originalModel)
	if err != nil {
		t.Fatalf("Failed to save model: %v", err)
	}
	
	loadedModel, err := LoadModel(filename)
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	
	device, _ := GetDeviceByID(loadedModel, "test-rtr-01")
	iface := device.Interfaces[0]
	
	if iface.Description != "Modified interface" {
		t.Errorf("Expected description 'Modified interface', got %s", iface.Description)
	}
	
	if iface.MTU != 8000 {
		t.Errorf("Expected MTU 8000, got %d", iface.MTU)
	}
}