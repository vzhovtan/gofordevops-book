package main

import (
	"flag"
	"fmt"
	"os"
	"securecom"
)

func main() {
	// Define command-line flags
	hostname := flag.String("hostname", "", "Server hostname or IP address")
	hostnames := flag.String("hostnames", "", "Comma-separated list of hostnames")
	port := flag.Int("port", 443, "HTTPS port (default: 443)")
	timeout := flag.Int("timeout", 10, "Request timeout in seconds")
	path := flag.String("path", "/", "URL path to check (default: /)")
	verifyTLS := flag.Bool("verify-tls", true, "Verify TLS certificates (default: true)")
	followRedirect := flag.Bool("follow-redirect", true, "Follow HTTP redirects (default: true)")
	continuous := flag.Bool("continuous", false, "Continuous monitoring mode")
	interval := flag.Int("interval", 60, "Check interval in seconds for continuous mode")

	flag.Parse()

	// Validate required flags
	if *hostname == "" && *hostnames == "" {
		fmt.Println("Error: either -hostname or -hostnames is required")
		flag.Usage()
		os.Exit(1)
	}

	securecom.ConnectAndCheck(hostname, hostnames, path, port, timeout, interval, verifyTLS, followRedirect, continuous)
}
