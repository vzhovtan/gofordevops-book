package configpush

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type DeploymentStrategy interface {
	Deploy(device *Device, config string) error
	Rollback(device *Device, backupConfig string) error
}

type ConfigBackup struct {
	DeviceID  string
	Timestamp time.Time
	Config    string
}

type FullReplaceStrategy struct {
	sshConfig *ssh.ClientConfig
	timeout   time.Duration
}

func NewFullReplaceStrategy(username, password string, timeout time.Duration) *FullReplaceStrategy {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	return &FullReplaceStrategy{
		sshConfig: config,
		timeout:   timeout,
	}
}

func (s *FullReplaceStrategy) connectSSH(device *Device) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:22", device.ManagementIP)
	client, err := ssh.Dial("tcp", addr, s.sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", device.ManagementIP, err)
	}
	return client, nil
}

func (s *FullReplaceStrategy) executeCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

func (s *FullReplaceStrategy) BackupCurrentConfig(device *Device) (*ConfigBackup, error) {
	client, err := s.connectSSH(device)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var backupCmd string
	switch device.Vendor {
	case "cisco":
		backupCmd = "show running-config"
	case "juniper":
		backupCmd = "show configuration"
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", device.Vendor)
	}

	config, err := s.executeCommand(client, backupCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to backup configuration: %w", err)
	}

	backup := &ConfigBackup{
		DeviceID:  device.ID,
		Timestamp: time.Now(),
		Config:    config,
	}

	return backup, nil
}

func (s *FullReplaceStrategy) deployCisco(client *ssh.Client, config string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	go func() {
		io.Copy(log.Writer(), stdout)
	}()

	commands := []string{
		"configure replace terminal\n",
		config,
		"\n",
		"end\n",
		"write memory\n",
		"exit\n",
	}

	for _, cmd := range commands {
		if _, err := stdin.Write([]byte(cmd)); err != nil {
			return fmt.Errorf("failed to write command: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err := session.Wait(); err != nil {
		return fmt.Errorf("session error: %w", err)
	}

	return nil
}

func (s *FullReplaceStrategy) Deploy(device *Device, config string) error {
	if device.Vendor != "cisco" {
		return fmt.Errorf("full replacement only supported for Cisco devices")
	}

	backup, err := s.BackupCurrentConfig(device)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	log.Printf("Created backup for device %s at %s", device.ID, backup.Timestamp.Format(time.RFC3339))

	client, err := s.connectSSH(device)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("Deploying full configuration to %s (%s)", device.Hostname, device.ManagementIP)

	if err := s.deployCisco(client, config); err != nil {
		log.Printf("Deployment failed, attempting rollback...")
		if rollbackErr := s.Rollback(device, backup.Config); rollbackErr != nil {
			return fmt.Errorf("deployment failed and rollback failed: deploy error: %w, rollback error: %v", err, rollbackErr)
		}
		return fmt.Errorf("deployment failed, successfully rolled back: %w", err)
	}

	if err := s.verifyConfiguration(device, config); err != nil {
		log.Printf("Configuration verification failed, rolling back...")
		if rollbackErr := s.Rollback(device, backup.Config); rollbackErr != nil {
			return fmt.Errorf("verification failed and rollback failed: %w", rollbackErr)
		}
		return fmt.Errorf("verification failed, rolled back: %w", err)
	}

	log.Printf("Configuration successfully deployed to %s", device.Hostname)
	return nil
}

func (s *FullReplaceStrategy) Rollback(device *Device, backupConfig string) error {
	client, err := s.connectSSH(device)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("Rolling back configuration on %s", device.Hostname)

	if err := s.deployCisco(client, backupConfig); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Printf("Configuration successfully rolled back on %s", device.Hostname)
	return nil
}

func (s *FullReplaceStrategy) verifyConfiguration(device *Device, expectedConfig string) error {
	time.Sleep(2 * time.Second)

	client, err := s.connectSSH(device)
	if err != nil {
		return err
	}
	defer client.Close()

	currentConfig, err := s.executeCommand(client, "show running-config")
	if err != nil {
		return fmt.Errorf("failed to retrieve current configuration: %w", err)
	}

	if !s.configMatches(currentConfig, expectedConfig) {
		return fmt.Errorf("configuration verification failed: deployed config does not match expected")
	}

	return nil
}

func (s *FullReplaceStrategy) configMatches(current, expected string) bool {
	currentLines := normalizeConfig(current)
	expectedLines := normalizeConfig(expected)

	criticalMatches := 0
	for _, line := range expectedLines {
		if isCriticalLine(line) {
			if containsLine(currentLines, line) {
				criticalMatches++
			}
		}
	}

	return criticalMatches > 0
}

func normalizeConfig(config string) []string {
	lines := strings.Split(config, "\n")
	var normalized []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "!") {
			continue
		}
		normalized = append(normalized, line)
	}

	return normalized
}

func isCriticalLine(line string) bool {
	criticalPrefixes := []string{
		"hostname",
		"interface",
		"ip address",
		"router",
		"network",
	}

	for _, prefix := range criticalPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func containsLine(lines []string, target string) bool {
	for _, line := range lines {
		if strings.Contains(line, target) {
			return true
		}
	}
	return false
}

type DeploymentResult struct {
	DeviceID  string
	Success   bool
	Error     error
	Duration  time.Duration
	Timestamp time.Time
	Backup    *ConfigBackup
}

type ConfigDeployer struct {
	strategy DeploymentStrategy
}

func NewConfigDeployer(strategy DeploymentStrategy) *ConfigDeployer {
	return &ConfigDeployer{
		strategy: strategy,
	}
}

func (d *ConfigDeployer) DeployToDevice(device *Device, config string) *DeploymentResult {
	startTime := time.Now()
	result := &DeploymentResult{
		DeviceID:  device.ID,
		Timestamp: startTime,
	}

	log.Printf("Starting deployment to device %s (%s)", device.ID, device.Hostname)

	err := d.strategy.Deploy(device, config)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = err
		log.Printf("Deployment failed for %s: %v", device.ID, err)
	} else {
		result.Success = true
		log.Printf("Deployment completed successfully for %s in %v", device.ID, result.Duration)
	}

	return result
}

func (d *ConfigDeployer) DeployToMultipleDevices(devices []*Device, configs map[string]string) []*DeploymentResult {
	results := make([]*DeploymentResult, 0, len(devices))

	for _, device := range devices {
		config, ok := configs[device.ID]
		if !ok {
			log.Printf("No configuration found for device %s, skipping", device.ID)
			continue
		}

		result := d.DeployToDevice(device, config)
		results = append(results, result)

		if !result.Success {
			log.Printf("Stopping deployment due to failure on %s", device.ID)
			break
		}
	}

	return results
}

func main() {
	model, err := LoadModel("infrastructure.json")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	device := &model.Devices[0]
	if device.Vendor != "cisco" {
		log.Fatalf("This example requires a Cisco device")
	}

	config, err := GenerateConfiguration(model, device.ID)
	if err != nil {
		log.Fatalf("Failed to generate configuration: %v", err)
	}

	strategy := NewFullReplaceStrategy("admin", "password", 30*time.Second)
	deployer := NewConfigDeployer(strategy)

	result := deployer.DeployToDevice(device, config)

	if result.Success {
		fmt.Printf("Deployment successful!\n")
		fmt.Printf("Device: %s\n", result.DeviceID)
		fmt.Printf("Duration: %v\n", result.Duration)
	} else {
		fmt.Printf("Deployment failed: %v\n", result.Error)
	}
}
