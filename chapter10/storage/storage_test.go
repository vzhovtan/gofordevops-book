package storage

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestNetworkDeviceStruct tests the NetworkDevice struct
func TestNetworkDeviceStruct(t *testing.T) {
	device := NetworkDevice{
		Hostname:    "test-router",
		IPAddress:   "10.0.0.1",
		DeviceType:  "Router",
		LastChecked: time.Now(),
		IsActive:    true,
	}

	if device.Hostname != "test-router" {
		t.Errorf("Expected hostname 'test-router', got '%s'", device.Hostname)
	}

	if device.IPAddress != "10.0.0.1" {
		t.Errorf("Expected IP '10.0.0.1', got '%s'", device.IPAddress)
	}

	if device.DeviceType != "Router" {
		t.Errorf("Expected device type 'Router', got '%s'", device.DeviceType)
	}

	if !device.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

// TestInventoryDataStruct tests the InventoryData struct
func TestInventoryDataStruct(t *testing.T) {
	devices := []NetworkDevice{
		{Hostname: "router-01", IPAddress: "192.168.1.1", DeviceType: "Router", IsActive: true},
		{Hostname: "switch-01", IPAddress: "192.168.1.2", DeviceType: "Switch", IsActive: true},
	}

	inventory := InventoryData{
		CollectionTime: time.Now(),
		TotalDevices:   len(devices),
		Devices:        devices,
	}

	if inventory.TotalDevices != 2 {
		t.Errorf("Expected 2 devices, got %d", inventory.TotalDevices)
	}

	if len(inventory.Devices) != 2 {
		t.Errorf("Expected 2 devices in slice, got %d", len(inventory.Devices))
	}
}

// TestSaveToJSON tests saving data to JSON file
func TestSaveToJSON(t *testing.T) {
	device := NetworkDevice{
		Hostname:    "test-device",
		IPAddress:   "172.16.0.1",
		DeviceType:  "Switch",
		LastChecked: time.Now(),
		IsActive:    true,
	}

	tmpfile, err := os.CreateTemp("", "test_save_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	err = SaveToJSON(tmpfile.Name(), device)
	if err != nil {
		t.Fatalf("SaveToJSON failed: %v", err)
	}

	// Verify the file exists and has content
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Saved file is empty")
	}

	// Verify JSON is valid
	var loadedDevice NetworkDevice
	err = json.Unmarshal(data, &loadedDevice)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved JSON: %v", err)
	}

	if loadedDevice.Hostname != device.Hostname {
		t.Errorf("Expected hostname '%s', got '%s'", device.Hostname, loadedDevice.Hostname)
	}
}

// TestSaveToJSON_InvalidPath tests error handling for invalid paths
func TestSaveToJSON_InvalidPath(t *testing.T) {
	device := NetworkDevice{
		Hostname:  "test",
		IPAddress: "1.2.3.4",
	}

	invalidPath := "/invalid/path/that/does/not/exist/file.json"
	err := SaveToJSON(invalidPath, device)

	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestSaveToJSON_InventoryData tests saving complete inventory
func TestSaveToJSON_InventoryData(t *testing.T) {
	devices := []NetworkDevice{
		{Hostname: "router-01", IPAddress: "192.168.1.1", DeviceType: "Router", IsActive: true},
		{Hostname: "switch-01", IPAddress: "192.168.1.2", DeviceType: "Switch", IsActive: true},
	}

	inventory := InventoryData{
		CollectionTime: time.Now(),
		TotalDevices:   len(devices),
		Devices:        devices,
	}

	tmpfile, err := os.CreateTemp("", "test_inventory_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	err = SaveToJSON(tmpfile.Name(), inventory)
	if err != nil {
		t.Fatalf("Failed to save inventory: %v", err)
	}

	// Load and verify
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var loadedInventory InventoryData
	err = json.Unmarshal(data, &loadedInventory)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if loadedInventory.TotalDevices != inventory.TotalDevices {
		t.Errorf("Expected %d devices, got %d", inventory.TotalDevices, loadedInventory.TotalDevices)
	}
}

// TestAppendDeviceToFile tests appending the device to th eexisting file
func TestAppendDeviceToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_append_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create initial inventory
	initialDevice := NetworkDevice{
		Hostname:    "device-01",
		IPAddress:   "192.168.1.1",
		DeviceType:  "Router",
		LastChecked: time.Now(),
		IsActive:    true,
	}

	inventory := InventoryData{
		CollectionTime: time.Now(),
		TotalDevices:   1,
		Devices:        []NetworkDevice{initialDevice},
	}

	err = SaveToJSON(tmpfile.Name(), inventory)
	if err != nil {
		t.Fatalf("Failed to save initial inventory: %v", err)
	}

	// Append a new device
	newDevice := NetworkDevice{
		Hostname:    "device-02",
		IPAddress:   "192.168.1.2",
		DeviceType:  "Switch",
		LastChecked: time.Now(),
		IsActive:    true,
	}

	err = AppendDeviceToFile(tmpfile.Name(), newDevice)
	if err != nil {
		t.Fatalf("Failed to append device: %v", err)
	}

	// Verify both devices exist
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var loadedInventory InventoryData
	err = json.Unmarshal(data, &loadedInventory)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if loadedInventory.TotalDevices != 2 {
		t.Errorf("Expected 2 devices, got %d", loadedInventory.TotalDevices)
	}

	if len(loadedInventory.Devices) != 2 {
		t.Errorf("Expected 2 devices in array, got %d", len(loadedInventory.Devices))
	}
}

// TestAppendDeviceToFile_NewFile tests appending to non-existent file
func TestAppendDeviceToFile_NewFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_new_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name()) // Remove to test creating new file

	device := NetworkDevice{
		Hostname:   "new-device",
		IPAddress:  "10.0.0.1",
		DeviceType: "Router",
		IsActive:   true,
	}

	err = AppendDeviceToFile(tmpfile.Name(), device)
	if err != nil {
		t.Fatalf("Failed to append to new file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Verify file was created
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read new file: %v", err)
	}

	var inventory InventoryData
	err = json.Unmarshal(data, &inventory)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if inventory.TotalDevices != 1 {
		t.Errorf("Expected 1 device, got %d", inventory.TotalDevices)
	}
}

// TestSaveDeviceMap tests saving devices as a map
func TestSaveDeviceMap(t *testing.T) {
	deviceMap := map[string]NetworkDevice{
		"router-01": {
			Hostname:   "router-01",
			IPAddress:  "192.168.1.1",
			DeviceType: "Router",
			IsActive:   true,
		},
		"switch-01": {
			Hostname:   "switch-01",
			IPAddress:  "192.168.1.2",
			DeviceType: "Switch",
			IsActive:   true,
		},
	}

	tmpfile, err := os.CreateTemp("", "test_map_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	err = SaveDeviceMap(tmpfile.Name(), deviceMap)
	if err != nil {
		t.Fatalf("Failed to save device map: %v", err)
	}

	// Verify map structure
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var loadedMap map[string]NetworkDevice
	err = json.Unmarshal(data, &loadedMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal map: %v", err)
	}

	if len(loadedMap) != 2 {
		t.Errorf("Expected 2 devices in map, got %d", len(loadedMap))
	}

	if _, exists := loadedMap["router-01"]; !exists {
		t.Error("Expected router-01 in map")
	}
}

// TestCreateBackup tests backup creation
func TestCreateBackup(t *testing.T) {
	// Create the original file
	tmpfile, err := os.CreateTemp("", "test_backup_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	originalName := tmpfile.Name()
	defer os.Remove(originalName)

	testData := []byte(`{"hostname":"test","ip":"1.2.3.4"}`)
	_, err = tmpfile.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpfile.Close()

	// Create backup
	err = CreateBackup(originalName)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Find the backup file (it has timestamp)
	dir := os.TempDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	backupFound := false
	for _, file := range files {
		if len(file.Name()) > len(originalName) &&
			file.Name()[:len(originalName)] == originalName[len(dir)+1:] {
			backupFound = true
			defer os.Remove(dir + "/" + file.Name())

			// Verify backup content
			backupData, err := os.ReadFile(dir + "/" + file.Name())
			if err != nil {
				t.Fatalf("Failed to read backup: %v", err)
			}

			if string(backupData) != string(testData) {
				t.Error("Backup content doesn't match original")
			}
		}
	}

	if !backupFound {
		t.Error("Backup file was not created")
	}
}

// TestCreateBackup_NonExistentFile tests backup error handling
func TestCreateBackup_NonExistentFile(t *testing.T) {
	err := CreateBackup("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestJSONMarshaling tests JSON marshaling of structs
func TestJSONMarshaling(t *testing.T) {
	device := NetworkDevice{
		Hostname:    "test-router",
		IPAddress:   "10.1.1.1",
		DeviceType:  "Router",
		LastChecked: time.Now(),
		IsActive:    true,
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal device: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled data is empty")
	}

	// Verify it can be unmarshaled back
	var unmarshaled NetworkDevice
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshaled.Hostname != device.Hostname {
		t.Error("Hostname mismatch after marshal/unmarshal")
	}
}
