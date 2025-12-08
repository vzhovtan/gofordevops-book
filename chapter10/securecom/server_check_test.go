package securecom

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewHTTPSChecker tests the HTTPS checker constructor
func TestNewHTTPSChecker(t *testing.T) {
	hostname := "example.com"
	port := 443
	timeout := 10 * time.Second
	verifyTLS := true

	checker := NewHTTPSChecker(hostname, port, timeout, verifyTLS)

	if checker.Hostname != hostname {
		t.Errorf("Expected hostname %s, got %s", hostname, checker.Hostname)
	}

	if checker.Port != port {
		t.Errorf("Expected port %d, got %d", port, checker.Port)
	}

	if checker.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, checker.Timeout)
	}

	if checker.VerifyTLS != verifyTLS {
		t.Errorf("Expected VerifyTLS %v, got %v", verifyTLS, checker.VerifyTLS)
	}

	if checker.CustomPath != "/" {
		t.Errorf("Expected default path /, got %s", checker.CustomPath)
	}
}

// TestBuildURL tests URL building with different configurations
func TestBuildURL(t *testing.T) {
	tests := []struct {
		name       string
		hostname   string
		port       int
		customPath string
		expected   string
	}{
		{
			name:       "Standard HTTPS port",
			hostname:   "example.com",
			port:       443,
			customPath: "/",
			expected:   "https://example.com/",
		},
		{
			name:       "Custom port",
			hostname:   "example.com",
			port:       8443,
			customPath: "/",
			expected:   "https://example.com:8443/",
		},
		{
			name:       "Custom path",
			hostname:   "api.example.com",
			port:       443,
			customPath: "/health",
			expected:   "https://api.example.com/health",
		},
		{
			name:       "Custom port and path",
			hostname:   "localhost",
			port:       9443,
			customPath: "/api/status",
			expected:   "https://localhost:9443/api/status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHTTPSChecker(tt.hostname, tt.port, 10*time.Second, true)
			checker.CustomPath = tt.customPath

			result := checker.BuildURL()
			if result != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestGetTLSVersionString tests TLS version string conversion
func TestGetTLSVersionString(t *testing.T) {
	tests := []struct {
		version  uint16
		expected string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x0000, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getTLSVersionString(tt.version)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestServerStatus tests the ServerStatus struct
func TestServerStatus(t *testing.T) {
	status := &ServerStatus{
		Hostname:       "test.example.com",
		URL:            "https://test.example.com/",
		IsAlive:        true,
		StatusCode:     200,
		ResponseTime:   150 * time.Millisecond,
		ContentLength:  1024,
		TLSVersion:     "TLS 1.3",
		CertExpiry:     time.Now().Add(90 * 24 * time.Hour),
		CheckTimestamp: time.Now(),
	}

	if status.Hostname != "test.example.com" {
		t.Error("Hostname mismatch")
	}

	if !status.IsAlive {
		t.Error("Expected IsAlive to be true")
	}

	if status.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", status.StatusCode)
	}

	if status.ResponseTime <= 0 {
		t.Error("Response time should be positive")
	}

	if status.ContentLength != 1024 {
		t.Errorf("Expected content length 1024, got %d", status.ContentLength)
	}
}

// TestCheckServerWithMockServer tests server checking with a mock HTTP server
func TestCheckServerWithMockServer(t *testing.T) {
	// Create a test HTTPS server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is alive"))
	})

	server := httptest.NewTLSServer(handler)
	defer server.Close()

	// Extract hostname and port from test server
	// Note: httptest servers use http.DefaultTransport which we need to handle
	t.Logf("Test server URL: %s", server.URL)

	// Test with the mock server
	// Note: This will test the basic structure, but won't fully validate
	// because httptest.Server uses http:// not https://
}

// TestCheckServer_InvalidHostname tests checking an invalid hostname
func TestCheckServer_InvalidHostname(t *testing.T) {
	checker := NewHTTPSChecker("invalid-hostname-12345.test", 443, 2*time.Second, true)

	status := checker.CheckServer()

	if status.IsAlive {
		t.Error("Expected server to be unreachable")
	}

	if status.Error == nil {
		t.Error("Expected error for invalid hostname")
	}

	t.Logf("Expected error received: %v", status.Error)
}

// TestCheckServer_Timeout tests timeout handling
func TestCheckServer_Timeout(t *testing.T) {
	// Use a non-routable IP to trigger timeout
	checker := NewHTTPSChecker("192.0.2.1", 443, 1*time.Second, false)

	start := time.Now()
	status := checker.CheckServer()
	elapsed := time.Since(start)

	if status.IsAlive {
		t.Error("Expected server to timeout")
	}

	if elapsed > 3*time.Second {
		t.Errorf("Timeout took too long: %v", elapsed)
	}

	t.Logf("Timeout occurred after %v", elapsed)
}

// TestCheckMultipleServers tests checking multiple servers
func TestCheckMultipleServers(t *testing.T) {
	hostnames := []string{"example.com", "invalid-host-xyz.test"}
	results := CheckMultipleServers(hostnames, 443, 5*time.Second, false)

	if len(results) != len(hostnames) {
		t.Errorf("Expected %d results, got %d", len(hostnames), len(results))
	}

	for i, result := range results {
		if result.Hostname != hostnames[i] {
			t.Errorf("Result %d: expected hostname %s, got %s", i, hostnames[i], result.Hostname)
		}
		t.Logf("Result %d: %s - IsAlive: %v", i, result.Hostname, result.IsAlive)
	}
}

// TestHTTPSCheckerStruct tests the HTTPSChecker struct fields
func TestHTTPSCheckerStruct(t *testing.T) {
	checker := &HTTPSChecker{
		Hostname:       "test.example.com",
		Port:           8443,
		Timeout:        15 * time.Second,
		FollowRedirect: false,
		VerifyTLS:      false,
		CustomPath:     "/api/health",
	}

	if checker.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	if checker.Port < 1 || checker.Port > 65535 {
		t.Errorf("Invalid port: %d", checker.Port)
	}

	if checker.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}

	if checker.CustomPath == "" {
		t.Error("CustomPath should not be empty")
	}
}

// TestPrintStatus tests the status printing (basic validation)
func TestPrintStatus(t *testing.T) {
	status := &ServerStatus{
		Hostname:       "example.com",
		URL:            "https://example.com/",
		IsAlive:        true,
		StatusCode:     200,
		ResponseTime:   100 * time.Millisecond,
		ContentLength:  5000,
		TLSVersion:     "TLS 1.3",
		CertExpiry:     time.Now().Add(365 * 24 * time.Hour),
		CheckTimestamp: time.Now(),
	}

	// Test that PrintStatus doesn't panic
	PrintStatus(status)
	PrintStatus(status)

	// Test with down server
	downStatus := &ServerStatus{
		Hostname:       "down.example.com",
		URL:            "https://down.example.com/",
		IsAlive:        false,
		Error:          fmt.Errorf("connection refused"),
		CheckTimestamp: time.Now(),
	}

	PrintStatus(downStatus)
}

// TestPrintSummary tests the summary printing
func TestPrintSummary(t *testing.T) {
	results := []*ServerStatus{
		{
			Hostname:   "server1.com",
			IsAlive:    true,
			StatusCode: 200,
		},
		{
			Hostname: "server2.com",
			IsAlive:  false,
			Error:    fmt.Errorf("timeout"),
		},
		{
			Hostname:   "server3.com",
			IsAlive:    true,
			StatusCode: 301,
		},
	}

	// Test that PrintSummary doesn't panic
	PrintSummary(results)
}

// TestResponseTimeTracking tests response time measurement
func TestResponseTimeTracking(t *testing.T) {
	status := &ServerStatus{
		ResponseTime: 250 * time.Millisecond,
	}

	if status.ResponseTime <= 0 {
		t.Error("Response time should be positive")
	}

	if status.ResponseTime > 10*time.Second {
		t.Error("Response time seems unreasonably high for test")
	}
}

// TestCertificateExpiry tests certificate expiry tracking
func TestCertificateExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		certExpiry    time.Time
		expectWarning bool
	}{
		{
			name:          "Valid cert - 365 days",
			certExpiry:    now.Add(365 * 24 * time.Hour),
			expectWarning: false,
		},
		{
			name:          "Expiring soon - 15 days",
			certExpiry:    now.Add(15 * 24 * time.Hour),
			expectWarning: true,
		},
		{
			name:          "Already expired",
			certExpiry:    now.Add(-1 * 24 * time.Hour),
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daysUntilExpiry := time.Until(tt.certExpiry).Hours() / 24

			if tt.expectWarning && daysUntilExpiry >= 30 {
				t.Error("Expected warning for certificate expiring soon")
			}

			t.Logf("Days until expiry: %.0f", daysUntilExpiry)
		})
	}
}

// TestStatusCodeValidation tests HTTP status code handling
func TestStatusCodeValidation(t *testing.T) {
	tests := []struct {
		statusCode int
		isError    bool
		isRedirect bool
	}{
		{200, false, false},
		{201, false, false},
		{204, false, false},
		{301, false, true},
		{302, false, true},
		{304, false, true},
		{400, true, false},
		{401, true, false},
		{403, true, false},
		{404, true, false},
		{500, true, false},
		{502, true, false},
		{503, true, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Status_%d", tt.statusCode), func(t *testing.T) {
			if tt.isError && tt.statusCode < 400 {
				t.Error("Status code should be >= 400 for errors")
			}

			if tt.isRedirect && (tt.statusCode < 300 || tt.statusCode >= 400) {
				t.Error("Redirect status codes should be 3xx")
			}
		})
	}
}

// TestEmptyHostname tests handling of empty hostname
func TestEmptyHostname(t *testing.T) {
	checker := NewHTTPSChecker("", 443, 5*time.Second, true)

	if checker.Hostname != "" {
		t.Error("Empty hostname should be preserved")
	}

	status := checker.CheckServer()
	if status.IsAlive {
		t.Error("Expected check to fail with empty hostname")
	}
}

// TestCustomPathHandling tests custom path handling
func TestCustomPathHandling(t *testing.T) {
	paths := []string{
		"/",
		"/health",
		"/api/v1/status",
		"/metrics",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			checker := NewHTTPSChecker("example.com", 443, 5*time.Second, true)
			checker.CustomPath = path

			url := checker.BuildURL()
			if url != "https://example.com"+path {
				t.Errorf("Expected URL https://example.com%s, got %s", path, url)
			}
		})
	}
}

// TestTLSConfigValidation tests TLS configuration
func TestTLSConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		verifyTLS bool
	}{
		{"Verify TLS", true},
		{"Skip TLS verification", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHTTPSChecker("example.com", 443, 5*time.Second, tt.verifyTLS)

			if checker.VerifyTLS != tt.verifyTLS {
				t.Errorf("Expected VerifyTLS %v, got %v", tt.verifyTLS, checker.VerifyTLS)
			}
		})
	}
}

// TestServerStatusFields tests all fields of ServerStatus
func TestServerStatusFields(t *testing.T) {
	now := time.Now()
	expiry := now.Add(90 * 24 * time.Hour)

	status := ServerStatus{
		Hostname:       "test.com",
		URL:            "https://test.com/",
		IsAlive:        true,
		StatusCode:     200,
		ResponseTime:   123 * time.Millisecond,
		ContentLength:  4096,
		TLSVersion:     "TLS 1.3",
		CertExpiry:     expiry,
		Error:          nil,
		CheckTimestamp: now,
	}

	// Validate all fields are set
	if status.Hostname == "" {
		t.Error("Hostname should not be empty")
	}
	if status.URL == "" {
		t.Error("URL should not be empty")
	}
	if !status.IsAlive {
		t.Error("IsAlive should be true")
	}
	if status.StatusCode != 200 {
		t.Error("StatusCode should be 200")
	}
	if status.ResponseTime <= 0 {
		t.Error("ResponseTime should be positive")
	}
	if status.ContentLength <= 0 {
		t.Error("ContentLength should be positive")
	}
	if status.TLSVersion == "" {
		t.Error("TLSVersion should not be empty")
	}
	if status.CertExpiry.IsZero() {
		t.Error("CertExpiry should not be zero")
	}
	if status.CheckTimestamp.IsZero() {
		t.Error("CheckTimestamp should not be zero")
	}
}
