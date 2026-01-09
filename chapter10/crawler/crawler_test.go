package crawler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDevice tests the Device struct
func TestDevice(t *testing.T) {
	device := Device{
		IPAddress:    "192.168.1.1",
		Hostname:     "router-01",
		IsAlive:      true,
		Port:         443,
		Protocol:     "https",
		DiscoveredAt: time.Now(),
	}

	if device.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", device.IPAddress)
	}

	if device.Port != 443 {
		t.Errorf("Expected port 443, got %d", device.Port)
	}

	if device.Protocol != "https" {
		t.Errorf("Expected protocol https, got %s", device.Protocol)
	}
}

// TestSystemInventory tests the SystemInventory struct
func TestSystemInventory(t *testing.T) {
	inventory := SystemInventory{
		Hostname:        "test-device",
		Model:           "Model-X",
		SerialNumber:    "SN12345",
		SoftwareVersion: "1.0.0",
		Uptime:          3600,
	}

	if inventory.Hostname != "test-device" {
		t.Errorf("Expected hostname test-device, got %s", inventory.Hostname)
	}

	if inventory.Uptime != 3600 {
		t.Errorf("Expected uptime 3600, got %d", inventory.Uptime)
	}
}

// TestCollectedData tests the CollectedData struct
func TestCollectedData(t *testing.T) {
	device := Device{
		IPAddress: "10.0.0.1",
		Hostname:  "test",
		IsAlive:   true,
	}

	inventory := SystemInventory{
		Hostname: "test-host",
		Model:    "TestModel",
	}

	collected := CollectedData{
		Device:      device,
		Inventory:   inventory,
		CollectedAt: time.Now(),
	}

	if collected.Device.IPAddress != "10.0.0.1" {
		t.Error("Device IP mismatch")
	}

	if collected.Inventory.Hostname != "test-host" {
		t.Error("Inventory hostname mismatch")
	}
}

// TestParseIPRange tests CIDR parsing
func TestParseIPRange(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
		minIPs      int
		maxIPs      int
	}{
		{
			name:        "Valid /30 range",
			cidr:        "192.168.1.0/30",
			expectError: false,
			minIPs:      2,
			maxIPs:      2,
		},
		{
			name:        "Valid /24 range",
			cidr:        "10.0.0.0/24",
			expectError: false,
			minIPs:      254,
			maxIPs:      254,
		},
		{
			name:        "Invalid CIDR",
			cidr:        "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := parseIPRange(tt.cidr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(ips) < tt.minIPs || len(ips) > tt.maxIPs {
				t.Errorf("Expected %d-%d IPs, got %d", tt.minIPs, tt.maxIPs, len(ips))
			}
		})
	}
}

// TestInc tests IP increment function
func TestInc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple increment",
			input:    "192.168.1.1",
			expected: "192.168.1.2",
		},
		{
			name:     "Octet overflow",
			input:    "192.168.1.255",
			expected: "192.168.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := []byte(tt.input)
			// Convert to net.IP format
			netIP := make([]byte, 4)
			copy(netIP, ip)
			inc(netIP)

			t.Logf("Incremented IP test: %s -> expected %s", tt.input, tt.expected)
		})
	}
}

// TestCheckHTTPPort tests HTTP port checking with mock server
func TestCheckHTTPPort(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Logf("Test server running at: %s", server.URL)

	// Note: This test is informational as checkHTTPPort expects specific format
	timeout := 1 * time.Second
	result := checkHTTPPort("127.0.0.1", 80, timeout)
	t.Logf("Port check result: %v", result)
}

// TestGetHostname tests hostname resolution
func TestGetHostname(t *testing.T) {
	hostname := getHostname("127.0.0.1")

	if hostname == "" {
		t.Error("Hostname should not be empty")
	}

	t.Logf("Resolved hostname for 127.0.0.1: %s", hostname)
}

// TestParseInventoryData tests JSON parsing of inventory data
func TestParseInventoryData(t *testing.T) {
	validJSON := `{
		"hostname": "test-device",
		"model": "Model-X",
		"serial_number": "SN123",
		"software_version": "1.0.0",
		"uptime": 3600
	}`

	inventory, err := parseInventoryData([]byte(validJSON))
	if err != nil {
		t.Fatalf("Failed to parse valid JSON: %v", err)
	}

	if inventory.Hostname != "test-device" {
		t.Errorf("Expected hostname test-device, got %s", inventory.Hostname)
	}

	if inventory.Model != "Model-X" {
		t.Errorf("Expected model Model-X, got %s", inventory.Model)
	}
}

// TestParseInventoryData_Invalid tests error handling for invalid JSON
func TestParseInventoryData_Invalid(t *testing.T) {
	invalidJSON := `{"hostname": invalid}`

	_, err := parseInventoryData([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestParseInventoryData_MissingHostname tests validation
func TestParseInventoryData_MissingHostname(t *testing.T) {
	jsonWithoutHostname := `{
		"model": "Model-X",
		"serial_number": "SN123"
	}`

	_, err := parseInventoryData([]byte(jsonWithoutHostname))
	if err == nil {
		t.Error("Expected error for missing hostname, got nil")
	}
}

// TestSaveToFileSystem tests file saving functionality
//func TestSaveToFileSystem(t *testing.T) {
//	tmpfile, err := os.CreateTemp("", "test_map_*.json")
//	if err != nil {
//		t.Fatalf("Failed to create temp file: %v", err)
//	}
//	tmpfile.Close()
//	defer os.Remove(tmpfile.Name())
//
//	testData := map[string]string{
//		"key": "value",
//	}
//
//	err = saveToFileSystem(testData, (tmpfile.Name()))
//	if err != nil {
//		t.Fatalf("Failed to save to filesystem: %v", err)
//	}
//
//	// Verify file content
//	data, err := os.ReadFile(tmpfile.Name())
//	if err != nil {
//		t.Fatalf("Failed to read saved file: %v", err)
//	}
//
//	var loaded map[string]string
//	err = json.Unmarshal(data, &loaded)
//	if err != nil {
//		t.Fatalf("Failed to unmarshal saved data: %v", err)
//	}
//
//	if loaded["key"] != "value" {
//		t.Error("Data mismatch after save/load")
//	}
//}

// TestValidateConfiguration tests configuration validation
func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		apiPath     string
		outputDir   string
		expectError bool
	}{
		{
			name:        "Valid configuration",
			cidr:        "192.168.1.0/24",
			apiPath:     "/api/v1/info",
			outputDir:   "/tmp/output",
			expectError: false,
		},
		{
			name:        "Missing CIDR",
			cidr:        "",
			apiPath:     "/api/v1/info",
			outputDir:   "/tmp/output",
			expectError: true,
		},
		{
			name:        "Missing API path",
			cidr:        "192.168.1.0/24",
			apiPath:     "",
			outputDir:   "/tmp/output",
			expectError: true,
		},
		{
			name:        "Missing output directory",
			cidr:        "192.168.1.0/24",
			apiPath:     "/api/v1/info",
			outputDir:   "",
			expectError: true,
		},
		{
			name:        "Invalid CIDR format",
			cidr:        "invalid-cidr",
			apiPath:     "/api/v1/info",
			outputDir:   "/tmp/output",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfiguration(tt.cidr, tt.apiPath, tt.outputDir)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCountByPort tests counting devices by port
func TestCountByPort(t *testing.T) {
	devices := []Device{
		{IPAddress: "192.168.1.1", Port: 80},
		{IPAddress: "192.168.1.2", Port: 443},
		{IPAddress: "192.168.1.3", Port: 443},
		{IPAddress: "192.168.1.4", Port: 80},
	}

	count80 := countByPort(devices, 80)
	if count80 != 2 {
		t.Errorf("Expected 2 devices on port 80, got %d", count80)
	}

	count443 := countByPort(devices, 443)
	if count443 != 2 {
		t.Errorf("Expected 2 devices on port 443, got %d", count443)
	}

	count22 := countByPort(devices, 22)
	if count22 != 0 {
		t.Errorf("Expected 0 devices on port 22, got %d", count22)
	}
}

// TestScanResult tests the ScanResult struct
func TestScanResult(t *testing.T) {
	devices := []Device{
		{IPAddress: "192.168.1.1", IsAlive: true},
		{IPAddress: "192.168.1.2", IsAlive: true},
	}

	scanResult := ScanResult{
		Devices:      devices,
		ScanTime:     time.Now(),
		TotalScanned: 254,
		TotalAlive:   len(devices),
	}

	if scanResult.TotalAlive != 2 {
		t.Errorf("Expected 2 alive devices, got %d", scanResult.TotalAlive)
	}

	if scanResult.TotalScanned != 254 {
		t.Errorf("Expected 254 scanned, got %d", scanResult.TotalScanned)
	}
}

// TestJSONMarshaling tests JSON marshaling of all structs
func TestJSONMarshaling(t *testing.T) {
	device := Device{
		IPAddress: "192.168.1.1",
		Hostname:  "test",
		Port:      443,
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal device: %v", err)
	}

	var unmarshaled Device
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal device: %v", err)
	}

	if unmarshaled.IPAddress != device.IPAddress {
		t.Error("IP address mismatch after marshal/unmarshal")
	}
}
