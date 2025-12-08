package main

import (
	"flag"
	"fmt"
	"os"
	"securecom"
)

//func main() {
//	//define CLI flags
//	hostname := flag.String("hostname", "", "Device hostname or IP address")
//	username := flag.String("username", "", "SSH username")
//	password := flag.String("password", "", "SSH password")
//	port := flag.Int("port", 22, "SSH port (default: 22)")
//	command := flag.String("command", "show configuration", "Command to execute")
//	timeout := flag.Int("timeout", 30, "Connection timeout in seconds")
//	outputFile := flag.String("output", "", "Save output to file (optional)")
//
//	flag.Parse()
//
//	// flags validation
//	if *hostname == "" {
//		fmt.Println("Error: hostname is required")
//		flag.Usage()
//		os.Exit(1)
//	}
//
//	if *username == "" {
//		fmt.Println("Error: username is required")
//		flag.Usage()
//		os.Exit(1)
//	}
//
//	if *password == "" {
//		fmt.Println("Error: password is required")
//		flag.Usage()
//		os.Exit(1)
//	}
//
//	securecom.ConnectAndRun(hostname, username, password, command, outputFile, port, timeout)
//}

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
