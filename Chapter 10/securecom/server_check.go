package securecom

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ServerStatus represents the status of a server check
type ServerStatus struct {
	Hostname       string
	URL            string
	IsAlive        bool
	StatusCode     int
	ResponseTime   time.Duration
	ContentLength  int64
	TLSVersion     string
	CertExpiry     time.Time
	Error          error
	CheckTimestamp time.Time
}

// HTTPSChecker handles HTTPS server validation
type HTTPSChecker struct {
	Hostname       string
	Port           int
	Timeout        time.Duration
	FollowRedirect bool
	VerifyTLS      bool
	CustomPath     string
}

// NewHTTPSChecker creates a new HTTPS checker instance
func NewHTTPSChecker(hostname string, port int, timeout time.Duration, verifyTLS bool) *HTTPSChecker {
	return &HTTPSChecker{
		Hostname:       hostname,
		Port:           port,
		Timeout:        timeout,
		FollowRedirect: true,
		VerifyTLS:      verifyTLS,
		CustomPath:     "/",
	}
}

// BuildURL constructs the full URL
func (c *HTTPSChecker) BuildURL() string {
	if c.Port == 443 {
		return fmt.Sprintf("https://%s%s", c.Hostname, c.CustomPath)
	}
	return fmt.Sprintf("https://%s:%d%s", c.Hostname, c.Port, c.CustomPath)
}

// CheckServer validates the server liveness
func (c *HTTPSChecker) CheckServer() *ServerStatus {
	status := &ServerStatus{
		Hostname:       c.Hostname,
		URL:            c.BuildURL(),
		CheckTimestamp: time.Now(),
	}

	// Create custom HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !c.VerifyTLS,
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{
		Timeout:   c.Timeout,
		Transport: transport,
	}

	// Disable redirect following if configured
	if !c.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Measure response time
	startTime := time.Now()

	resp, err := client.Get(status.URL)
	status.ResponseTime = time.Since(startTime)

	if err != nil {
		status.IsAlive = false
		status.Error = err
		return status
	}
	defer resp.Body.Close()

	// Server is alive
	status.IsAlive = true
	status.StatusCode = resp.StatusCode

	// Read response body to get content length
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		status.ContentLength = int64(len(body))
	}

	// Get TLS information
	if resp.TLS != nil {
		status.TLSVersion = getTLSVersionString(resp.TLS.Version)
		if len(resp.TLS.PeerCertificates) > 0 {
			status.CertExpiry = resp.TLS.PeerCertificates[0].NotAfter
		}
	}

	return status
}

// getTLSVersionString converts TLS version to string
func getTLSVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

// PrintStatus prints the server status in a formatted way
func PrintStatus(status *ServerStatus) {
	fmt.Println("HTTPS Server Liveness Check")
	fmt.Printf("Hostname:        %s\n", status.Hostname)
	fmt.Printf("URL:             %s\n", status.URL)
	fmt.Printf("Check Time:      %s\n", status.CheckTimestamp.Format("2006-01-02 15:04:05"))

	if status.IsAlive {
		fmt.Printf("Status:          ✓ ALIVE\n")
		fmt.Printf("HTTP Code:       %d %s\n", status.StatusCode, http.StatusText(status.StatusCode))
		fmt.Printf("Response Time:   %v\n", status.ResponseTime)
		fmt.Printf("Content Length:  %d bytes\n", status.ContentLength)

		// Status code warnings
		if status.StatusCode >= 400 {
			fmt.Printf("\n⚠ WARNING: Server returned error status code %d\n", status.StatusCode)
		} else if status.StatusCode >= 300 && status.StatusCode < 400 {
			fmt.Printf("\nℹ INFO: Server returned redirect status code %d\n", status.StatusCode)
		}
	} else {
		fmt.Printf("Status:          ✗ DOWN/UNREACHABLE\n")
		fmt.Printf("Error:           %v\n", status.Error)
	}

}

// CheckMultipleServers checks multiple hostnames
func CheckMultipleServers(hostnames []string, port int, timeout time.Duration, verifyTLS bool) []*ServerStatus {
	results := make([]*ServerStatus, 0, len(hostnames))

	fmt.Printf("\nChecking %d server(s)...\n", len(hostnames))

	for i, hostname := range hostnames {
		fmt.Printf("\n[%d/%d] Checking %s...\n", i+1, len(hostnames), hostname)
		checker := NewHTTPSChecker(hostname, port, timeout, verifyTLS)
		status := checker.CheckServer()
		results = append(results, status)
	}

	return results
}

// PrintSummary prints a summary of multiple checks
func PrintSummary(results []*ServerStatus) {
	fmt.Println("Summary")

	alive := 0
	down := 0

	for _, result := range results {
		if result.IsAlive {
			alive++
			fmt.Printf("✓ %-30s [%d] %v\n", result.Hostname, result.StatusCode, result.ResponseTime)
		} else {
			down++
			fmt.Printf("✗ %-30s [ERROR]\n", result.Hostname)
		}
	}

	fmt.Printf("Total Servers:   %d\n", len(results))
	fmt.Printf("Alive:           %d\n", alive)
	fmt.Printf("Down:            %d\n", down)
}

func ConnectAndCheck(hostname, hostnames, path *string, port, timeout, interval *int, verifyTLS, followRedirect, continuous *bool) {
	// Parse hostnames
	var hostList []string
	if *hostname != "" {
		hostList = append(hostList, *hostname)
	}

	if *hostnames != "" {
		// Simple split by comma (could be improved with proper CSV parsing)
		for i := 0; i < len(*hostnames); i++ {
			start := i
			for i < len(*hostnames) && (*hostnames)[i] != ',' {
				i++
			}
			host := (*hostnames)[start:i]
			// Trim spaces
			for len(host) > 0 && host[0] == ' ' {
				host = host[1:]
			}
			for len(host) > 0 && host[len(host)-1] == ' ' {
				host = host[:len(host)-1]
			}
			if len(host) > 0 {
				hostList = append(hostList, host)
			}
		}
	}

	// Continuous monitoring mode
	if *continuous {
		fmt.Printf("Starting continuous monitoring (interval: %d seconds)\n", *interval)
		fmt.Println("Press Ctrl+C to stop")

		for {
			results := CheckMultipleServers(hostList, *port, time.Duration(*timeout)*time.Second, *verifyTLS)

			if len(results) == 1 {
				PrintStatus(results[0])
			} else {
				PrintSummary(results)
			}

			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}

	// Single check mode
	results := CheckMultipleServers(hostList, *port, time.Duration(*timeout)*time.Second, *verifyTLS)

	if len(results) == 1 {
		// Single server - detailed output
		checker := NewHTTPSChecker(hostList[0], *port, time.Duration(*timeout)*time.Second, *verifyTLS)
		checker.CustomPath = *path
		checker.FollowRedirect = *followRedirect

		status := checker.CheckServer()
		PrintStatus(status)

		// Exit with appropriate code
		if !status.IsAlive {
			os.Exit(1)
		}
	} else {
		// Multiple servers - summary output
		PrintSummary(results)

		// Exit with error if any server is down
		for _, result := range results {
			if !result.IsAlive {
				os.Exit(1)
			}
		}
	}

	fmt.Println("\nCheck completed successfully!")
}
