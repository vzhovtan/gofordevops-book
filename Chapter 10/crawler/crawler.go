package crawler

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Device represents a discovered infrastructure device
type Device struct {
	IPAddress    string    `json:"ip_address"`
	Hostname     string    `json:"hostname"`
	IsAlive      bool      `json:"is_alive"`
	Port         int       `json:"port"`
	Protocol     string    `json:"protocol"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

// SystemInventory represents API response from device
type SystemInventory struct {
	Hostname        string `json:"hostname"`
	Model           string `json:"model"`
	SerialNumber    string `json:"serial_number"`
	SoftwareVersion string `json:"software_version"`
	Uptime          int    `json:"uptime"`
}

// CollectedData represents the complete collected information
type CollectedData struct {
	Device      Device          `json:"device"`
	Inventory   SystemInventory `json:"inventory"`
	CollectedAt time.Time       `json:"collected_at"`
}

// ScanResult holds the results from network scanning
type ScanResult struct {
	Devices      []Device  `json:"devices"`
	ScanTime     time.Time `json:"scan_time"`
	TotalScanned int       `json:"total_scanned"`
	TotalAlive   int       `json:"total_alive"`
}

// parseIPRange parses CIDR and returns list of IP addresses
func parseIPRange(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIDR: %v", err)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast addresses
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// checkHTTPPort validates if HTTP/HTTPS port is open and responsive
func checkHTTPPort(ip string, port int, timeout time.Duration) bool {
	protocol := "http"
	if port == 443 {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s:%d/", protocol, ip, port)

	// Create custom HTTP client with timeout and TLS config
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip verification for discovery
		},
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode > 0
}

// getHostname performs reverse DNS lookup
func getHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "unknown"
	}
	return names[0]
}

// scanNetworkForAPI scans network range for devices with API endpoints
func scanNetworkForAPI(cidr string, workers int, timeout time.Duration) ([]Device, error) {
	ips, err := parseIPRange(cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP range: %v", err)
	}

	fmt.Printf("Scanning %d IP addresses for API endpoints...\n", len(ips))

	var wg sync.WaitGroup
	ipChan := make(chan string, len(ips))
	resultChan := make(chan Device, len(ips))

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				// Check both HTTP and HTTPS ports
				for _, port := range []int{80, 443} {
					if checkHTTPPort(ip, port, timeout) {
						protocol := "http"
						if port == 443 {
							protocol = "https"
						}

						device := Device{
							IPAddress:    ip,
							Hostname:     getHostname(ip),
							IsAlive:      true,
							Port:         port,
							Protocol:     protocol,
							DiscoveredAt: time.Now(),
						}
						resultChan <- device
						fmt.Printf("Found API endpoint: %s:%d (%s)\n", ip, port, protocol)
						break
					}
				}
			}
		}()
	}

	// Send IPs to workers
	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var devices []Device
	for device := range resultChan {
		devices = append(devices, device)
	}

	fmt.Printf("Scan complete. Found %d API endpoints.\n", len(devices))
	return devices, nil
}

// retrieveInventoryFromAPI fetches system inventory from device API
func retrieveInventoryFromAPI(device Device, apiPath string, timeout time.Duration) (*SystemInventory, error) {
	url := fmt.Sprintf("%s://%s:%d%s", device.Protocol, device.IPAddress, device.Port, apiPath)

	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	fmt.Printf("Retrieving inventory from %s...\n", url)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var inventory SystemInventory
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &inventory, nil
}

// parseInventoryData parses JSON inventory data into Go structure
func parseInventoryData(jsonData []byte) (*SystemInventory, error) {
	var inventory SystemInventory
	err := json.Unmarshal(jsonData, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory JSON: %v", err)
	}

	// Validate required fields
	if inventory.Hostname == "" {
		return nil, fmt.Errorf("missing required field: hostname")
	}

	return &inventory, nil
}

// saveToFileSystem stores collected data to file system with timestamp
func saveToFileSystem(data interface{}, baseFilename string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", baseFilename, timestamp)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Data saved to: %s\n", filename)
	return nil
}

// saveCollectedData saves individual device collected data
func saveCollectedData(device Device, inventory *SystemInventory, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	collected := CollectedData{
		Device:      device,
		Inventory:   *inventory,
		CollectedAt: time.Now(),
	}

	// Create filename based on IP address
	safeIP := device.IPAddress
	for i := 0; i < len(safeIP); i++ {
		if safeIP[i] == '.' {
			safeIP = safeIP[:i] + "_" + safeIP[i+1:]
		}
	}

	filename := fmt.Sprintf("%s/device_%s", outputDir, safeIP)
	return saveToFileSystem(collected, filename)
}

// saveScanResults saves the complete scan results
func saveScanResults(devices []Device, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	scanResult := ScanResult{
		Devices:      devices,
		ScanTime:     time.Now(),
		TotalScanned: 0, // Can be calculated from CIDR
		TotalAlive:   len(devices),
	}

	filename := fmt.Sprintf("%s/scan_results", outputDir)
	return saveToFileSystem(scanResult, filename)
}

// collectInventoryFromDevices retrieves and stores inventory from all discovered devices
func collectInventoryFromDevices(devices []Device, apiPath string, timeout time.Duration, outputDir string) {
	fmt.Printf("\nCollecting inventory from %d devices...\n", len(devices))

	for i, device := range devices {
		fmt.Printf("\n[%d/%d] Processing %s:%d...\n", i+1, len(devices), device.IPAddress, device.Port)

		inventory, err := retrieveInventoryFromAPI(device, apiPath, timeout)
		if err != nil {
			log.Printf("Failed to retrieve inventory from %s: %v", device.IPAddress, err)
			continue
		}

		fmt.Printf("Retrieved inventory: Hostname=%s, Model=%s, Serial=%s\n",
			inventory.Hostname, inventory.Model, inventory.SerialNumber)

		err = saveCollectedData(device, inventory, outputDir)
		if err != nil {
			log.Printf("Failed to save data for %s: %v", device.IPAddress, err)
			continue
		}

		fmt.Printf("Successfully saved inventory for %s\n", device.IPAddress)
	}
}

// validateConfiguration checks if configuration is valid
func validateConfiguration(cidr, apiPath, outputDir string) error {
	if cidr == "" {
		return fmt.Errorf("CIDR range is required")
	}

	if apiPath == "" {
		return fmt.Errorf("API path is required")
	}

	if outputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	// Validate CIDR format
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %v", err)
	}

	return nil
}

// printSummary displays collection summary
func printSummary(devices []Device, startTime time.Time) {
	fmt.Println("Collection Summary")
	fmt.Printf("Total devices discovered: %d\n", len(devices))
	fmt.Printf("HTTP endpoints:           %d\n", countByPort(devices, 80))
	fmt.Printf("HTTPS endpoints:          %d\n", countByPort(devices, 443))
	fmt.Printf("Total time elapsed:       %v\n", time.Since(startTime))
}

// countByPort counts devices by port number
func countByPort(devices []Device, port int) int {
	count := 0
	for _, device := range devices {
		if device.Port == port {
			count++
		}
	}
	return count
}

func Crawl() {
	// Configuration example
	cidr := "192.168.1.0/24"         // Network range to scan
	apiPath := "/api/v1/system/info" // API endpoint path
	outputDir := "collected_data"    // Output directory
	workers := 50                    // Concurrent workers
	timeout := 2 * time.Second       // Request timeout

	fmt.Println("Network API Discovery and Inventory Collection")

	// Validate configuration
	err := validateConfiguration(cidr, apiPath, outputDir)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Printf("CIDR Range:    %s\n", cidr)
	fmt.Printf("API Path:      %s\n", apiPath)
	fmt.Printf("Output Dir:    %s\n", outputDir)
	fmt.Printf("Workers:       %d\n", workers)
	fmt.Printf("Timeout:       %v\n", timeout)

	startTime := time.Now()

	// Step 1: Scan network for API endpoints (ports 80 and 443)
	fmt.Println("\nStep 1: Scanning network for API endpoints...")
	devices, err := scanNetworkForAPI(cidr, workers, timeout)
	if err != nil {
		log.Fatalf("Network scan failed: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No API endpoints found. Exiting.")
		return
	}

	// Step 2: Save scan results
	fmt.Println("\nStep 2: Saving scan results...")
	err = saveScanResults(devices, outputDir)
	if err != nil {
		log.Printf("Warning: Failed to save scan results: %v", err)
	}

	// Step 3: Collect inventory from discovered devices
	fmt.Println("\nStep 3: Collecting inventory from devices...")
	collectInventoryFromDevices(devices, apiPath, timeout*2, outputDir)

	// Step 4: Print summary
	printSummary(devices, startTime)

	fmt.Println("\nCollection complete! Check output directory:", outputDir)
}
