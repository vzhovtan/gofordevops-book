package discovery

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// ScannedDevice represents a discovered network device
type ScannedDevice struct {
	IPAddress string
	Hostname  string
	IsAlive   bool
}

// parseIPRange parses CIDR notation and returns list of IP addresses
func parseIPRange(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
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

// pingHost checks if a host is reachable attempting TCP connection
func pingHost(ip string, timeout time.Duration) bool {
	// Try common ports for network devices and servers to check if the host is alive
	ports := []string{"80", "443", "22", "23"}

	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

// getHostname performs reverse DNS lookup to get hostname
func getHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "Unknown"
	}
	return names[0]
}

// scanIP scans a single IP address
func scanIP(ip string, timeout time.Duration) ScannedDevice {
	device := ScannedDevice{
		IPAddress: ip,
		IsAlive:   false,
	}

	// Check if the host is alive
	if pingHost(ip, timeout) {
		device.IsAlive = true
		device.Hostname = getHostname(ip)
	}

	return device
}

// scanNetwork scans the entire network range
func scanNetwork(cidr string, workers int, timeout time.Duration) []ScannedDevice {
	ips, err := parseIPRange(cidr)
	if err != nil {
		fmt.Printf("Error parsing IP range: %v\n", err)
		return nil
	}

	fmt.Printf("Scanning %d IP addresses in range %s\n", len(ips), cidr)
	fmt.Printf("Using %d workers with %v timeout per host\n\n", workers, timeout)

	var wg sync.WaitGroup
	ipChan := make(chan string, len(ips))
	resultChan := make(chan ScannedDevice, len(ips))

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				device := scanIP(ip, timeout)
				resultChan <- device
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
	var devices []ScannedDevice
	scanned := 0
	for device := range resultChan {
		scanned++
		if device.IsAlive {
			devices = append(devices, device)
			fmt.Printf("[%d/%d] Found: %s -> %s\n", scanned, len(ips), device.IPAddress, device.Hostname)
		} else {
			if scanned%10 == 0 {
				fmt.Printf("[%d/%d] Scanning...\n", scanned, len(ips))
			}
		}
	}

	return devices
}

func Scanner() {
	// Default configuration
	cidr := "192.168.1.0/24"
	workers := 50
	timeout := 500 * time.Millisecond

	// Parse command line arguments
	if len(os.Args) > 1 {
		cidr = os.Args[1]
	}

	fmt.Println("=== Network Scanner ===")
	fmt.Printf("Target: %s\n\n", cidr)

	startTime := time.Now()

	// Scan the network
	devices := scanNetwork(cidr, workers, timeout)

	elapsed := time.Since(startTime)

	// Print results
	fmt.Printf("\n=== Scan Complete ===\n")
	fmt.Printf("Time elapsed: %v\n", elapsed)
	fmt.Printf("Active devices found: %d\n\n", len(devices))

	if len(devices) > 0 {
		fmt.Println("Active Devices:")
		fmt.Println("----------------------------------------")
		for i, device := range devices {
			fmt.Printf("%2d. %-15s -> %s\n", i+1, device.IPAddress, device.Hostname)
		}
	} else {
		fmt.Println("No active devices found in the specified range.")
	}
}
