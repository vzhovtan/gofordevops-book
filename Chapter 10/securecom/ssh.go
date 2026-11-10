package securecom

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"time"
)

// Struct SSHClient represents an SSH connection configuration
type SSHClient struct {
	Hostname string
	Username string
	Password string
	Port     int
	Timeout  time.Duration
}

// NewSSHClient creates a new SSH client instance and returns the pointer to SSHClient
func NewSSHClient(hostname, username, password string, port int, timeout time.Duration) *SSHClient {
	return &SSHClient{
		Hostname: hostname,
		Username: username,
		Password: password,
		Port:     port,
		Timeout:  timeout,
	}
}

// Method Connect establishes SSH connection and returns the client
// Note: Use proper host key verification in production
func (c *SSHClient) Connect() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.Timeout,
	}

	address := fmt.Sprintf("%s:%d", c.Hostname, c.Port)
	fmt.Printf("Connecting to %s@%s...\n", c.Username, address)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	fmt.Println("Connection established successfully!")
	return client, nil
}

// Method ExecuteCommand executes a single command on the remote device
func (c *SSHClient) ExecuteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	fmt.Printf("Executing command: %s\n", command)

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	return string(output), nil
}

// Method ExecuteCommands executes multiple commands on the remote device
func (c *SSHClient) ExecuteCommands(client *ssh.Client, commands []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, cmd := range commands {
		output, err := c.ExecuteCommand(client, cmd)
		if err != nil {
			return results, err
		}
		results[cmd] = output
	}

	return results, nil
}

// SaveToFile saves the output to a file
func SaveToFile(filename, content string) error {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}
	fmt.Printf("Output saved to: %s\n", filename)
	return nil
}

func ConnectAndRun(hostname, username, password, command, outputFile *string, port, timeout *int) {
	// Create SSH client
	sshClient := NewSSHClient(*hostname, *username, *password, *port, time.Duration(*timeout)*time.Second)

	// Connect to device
	client, err := sshClient.Connect()
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Close()

	// Execute command
	output, err := sshClient.ExecuteCommand(client, *command)
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}

	fmt.Println("Command Output:")
	fmt.Println(output)

	// Save to file if specified
	if *outputFile != "" {
		err := SaveToFile(*outputFile, output)
		if err != nil {
			log.Printf("Warning: %v", err)
		}
	}
}
