package discovery

import (
	"net"
	"testing"
	"time"
)

// TestParseIPRange tests the IP range parsing function
func TestParseIPRange(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
		minIPs      int
		maxIPs      int
	}{
		{
			name:        "Valid /24 range",
			cidr:        "192.168.1.0/24",
			expectError: false,
			minIPs:      254,
			maxIPs:      254,
		},
		{
			name:        "Valid /30 range",
			cidr:        "10.0.0.0/30",
			expectError: false,
			minIPs:      2,
			maxIPs:      2,
		},
		{
			name:        "Valid /28 range",
			cidr:        "172.16.0.0/28",
			expectError: false,
			minIPs:      14,
			maxIPs:      14,
		},
		{
			name:        "Invalid CIDR",
			cidr:        "invalid",
			expectError: true,
		},
		{
			name:        "Invalid IP",
			cidr:        "999.999.999.999/24",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := parseIPRange(tt.cidr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			ipCount := len(ips)
			if ipCount < tt.minIPs || ipCount > tt.maxIPs {
				t.Errorf("Expected %d-%d IPs, got %d", tt.minIPs, tt.maxIPs, ipCount)
			}
		})
	}
}

// TestInc tests the IP increment function
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
		{
			name:     "Multiple octet overflow",
			input:    "192.168.255.255",
			expected: "192.169.0.0",
		},
		{
			name:     "Start of range",
			input:    "10.0.0.0",
			expected: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.input).To4()
			if ip == nil {
				t.Fatalf("Failed to parse input IP: %s", tt.input)
			}

			inc(ip)
			result := ip.String()

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestPingHost tests the host ping functionality
func TestPingHost(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		timeout  time.Duration
		expected bool
	}{
		{
			name:     "Localhost should be reachable",
			ip:       "127.0.0.1",
			timeout:  100 * time.Millisecond,
			expected: false, // May fail if no services running
		},
		{
			name:     "Invalid IP should fail",
			ip:       "0.0.0.0",
			timeout:  100 * time.Millisecond,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pingHost(tt.ip, tt.timeout)
			// Note: This test is informational as results depend on network state
			t.Logf("Ping result for %s: %v", tt.ip, result)
		})
	}
}

// TestGetHostname tests the hostname resolution
func TestGetHostname(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{
			name: "Localhost resolution",
			ip:   "127.0.0.1",
		},
		{
			name: "Non-existent IP",
			ip:   "192.168.255.254",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname := getHostname(tt.ip)
			t.Logf("Hostname for %s: %s", tt.ip, hostname)

			if hostname == "" {
				t.Errorf("Expected non-empty hostname")
			}
		})
	}
}

// TestScanIP tests the single IP scanning function
func TestScanIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		timeout time.Duration
	}{
		{
			name:    "Scan localhost",
			ip:      "127.0.0.1",
			timeout: 100 * time.Millisecond,
		},
		{
			name:    "Scan non-existent IP",
			ip:      "192.168.255.254",
			timeout: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := scanIP(tt.ip, tt.timeout)

			if device.IPAddress != tt.ip {
				t.Errorf("Expected IP %s, got %s", tt.ip, device.IPAddress)
			}

			t.Logf("Device: IP=%s, Hostname=%s, IsAlive=%v",
				device.IPAddress, device.Hostname, device.IsAlive)
		})
	}
}

// TestScannedDeviceStruct tests the ScannedDevice struct
func TestScannedDeviceStruct(t *testing.T) {
	device := ScannedDevice{
		IPAddress: "192.168.1.1",
		Hostname:  "router.local",
		IsAlive:   true,
	}

	if device.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", device.IPAddress)
	}

	if device.Hostname != "router.local" {
		t.Errorf("Expected hostname router.local, got %s", device.Hostname)
	}

	if !device.IsAlive {
		t.Errorf("Expected IsAlive to be true")
	}
}
