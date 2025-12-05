package discovery

import (
	"encoding/json"
	"os"
	"testing"
)

// TestDeviceStruct tests the Device struct
func TestDeviceStruct(t *testing.T) {
	device := Device{
		Hostname:    "test-router",
		IPv4Address: "192.168.1.1",
	}

	if device.Hostname != "test-router" {
		t.Errorf("Expected hostname 'test-router', got '%s'", device.Hostname)
	}

	if device.IPv4Address != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", device.IPv4Address)
	}
}

// TestNetworkDevicesStruct tests the NetworkDevices struct
func TestNetworkDevicesStruct(t *testing.T) {
	devices := []Device{
		{Hostname: "router-01", IPv4Address: "192.168.1.1"},
		{Hostname: "switch-01", IPv4Address: "192.168.1.2"},
	}

	infraDevices := InfraDevices{
		Devices: devices,
	}

	if len(infraDevices.Devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(infraDevices.Devices))
	}

	if infraDevices.Devices[0].Hostname != "router-01" {
		t.Errorf("Expected first device hostname 'router-01', got '%s'",
			infraDevices.Devices[0].Hostname)
	}
}

// TestJSONMarshaling tests JSON marshaling of Device struct
func TestJSONMarshaling(t *testing.T) {
	device := Device{
		Hostname:    "test-device",
		IPv4Address: "10.0.0.1",
	}

	jsonData, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal device: %v", err)
	}

	expectedJSON := `{"hostname":"test-device","ipv4_address":"10.0.0.1"}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Expected JSON: %s, got: %s", expectedJSON, string(jsonData))
	}
}

// TestJSONUnmarshaling tests JSON unmarshaling into Device struct
func TestJSONUnmarshaling(t *testing.T) {
	jsonData := []byte(`{"hostname":"test-device","ipv4_address":"10.0.0.1"}`)

	var device Device
	err := json.Unmarshal(jsonData, &device)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if device.Hostname != "test-device" {
		t.Errorf("Expected hostname 'test-device', got '%s'", device.Hostname)
	}

	if device.IPv4Address != "10.0.0.1" {
		t.Errorf("Expected IP '10.0.0.1', got '%s'", device.IPv4Address)
	}
}

// TestReadDevicesFromFile tests reading devices from a valid JSON file
func TestReadDevicesFromFile(t *testing.T) {
	// Create the temporary test file
	testData := `{
  "infrastructure_devices": [
    {
      "hostname": "router-test",
      "ipv4_address": "192.168.1.1"
    },
    {
      "hostname": "switch-test",
      "ipv4_address": "192.168.1.2"
    }
  ]
}`

	tmpfile, err := os.CreateTemp("", "test_devices_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test reading the file
	devices, err := ReadDevicesFromFile(tmpfile.Name())

	if err != nil {
		t.Fatalf("Failed to read devices from file: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}

	if devices[0].Hostname != "router-test" {
		t.Errorf("Expected first device hostname 'router-test', got '%s'",
			devices[0].Hostname)
	}

	if devices[1].IPv4Address != "192.168.1.2" {
		t.Errorf("Expected second device IP '192.168.1.2', got '%s'",
			devices[1].IPv4Address)
	}
}

// TestReadDevicesFromFile_NonExistentFile tests error handling for missing file
func TestReadDevicesFromFile_NonExistentFile(t *testing.T) {
	_, err := ReadDevicesFromFile("non_existent_file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestReadDevicesFromFile_InvalidJSON tests error handling for invalid JSON
func TestReadDevicesFromFile_InvalidJSON(t *testing.T) {
	// Create the temporary file with invalid JSON
	invalidJSON := `{
  "infrastructure_devices": [
    {
      "hostname": "router-test"
      "ipv4_address": "192.168.1.1"
    }
  ]
}`

	tmpfile, err := os.CreateTemp("", "invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(invalidJSON)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test reading the invalid file
	_, err = ReadDevicesFromFile(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestReadDevicesFromFile_EmptyDevicesList tests handling of empty devices list
func TestReadDevicesFromFile_EmptyDevicesList(t *testing.T) {
	emptyData := `{"infrastructure_devices": []}`

	tmpfile, err := os.CreateTemp("", "empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(emptyData)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	devices, err := ReadDevicesFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read devices from file: %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices, got %d", len(devices))
	}
}

// TestReadDevicesFromFile_LargeFile tests handling of files with many devices
func TestReadDevicesFromFile_LargeFile(t *testing.T) {
	// Create a JSON file with 500 devices
	var infraDevices InfraDevices
	for i := 1; i <= 500; i++ {
		device := Device{
			Hostname:    "device-" + string(rune(i)),
			IPv4Address: "192.168.1." + string(rune(i)),
		}
		infraDevices.Devices = append(infraDevices.Devices, device)
	}

	jsonData, err := json.MarshalIndent(infraDevices, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal devices: %v", err)
	}

	tmpfile, err := os.CreateTemp("", "large_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	devices, err := ReadDevicesFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read devices from file: %v", err)
	}

	if len(devices) != 500 {
		t.Errorf("Expected 500 devices, got %d", len(devices))
	}
}

// TestDeviceJSONTags tests that JSON tags are correctly defined
func TestDeviceJSONTags(t *testing.T) {
	jsonStr := `{"hostname":"test","ipv4_address":"1.2.3.4"}`
	var device Device

	err := json.Unmarshal([]byte(jsonStr), &device)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if device.Hostname != "test" {
		t.Errorf("Hostname not properly unmarshaled")
	}

	if device.IPv4Address != "1.2.3.4" {
		t.Errorf("IPv4Address not properly unmarshaled")
	}

	// Test marshaling back
	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if string(data) != jsonStr {
		t.Errorf("Expected %s, got %s", jsonStr, string(data))
	}
}
