package infra

import (
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type ConfigElement struct {
	Type        string
	Path        string
	Operation   string
	Value       string
	Description string
}

type PerElementStrategy struct {
	sshConfig *ssh.ClientConfig
	timeout   time.Duration
}

func NewPerElementStrategy(username, password string, timeout time.Duration) *PerElementStrategy {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	return &PerElementStrategy{
		sshConfig: config,
		timeout:   timeout,
	}
}

func (s *PerElementStrategy) connectSSH(device *Device) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:22", device.ManagementIP)
	client, err := ssh.Dial("tcp", addr, s.sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", device.ManagementIP, err)
	}
	return client, nil
}

func (s *PerElementStrategy) executeCommands(client *ssh.Client, commands []string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return "", fmt.Errorf("failed to request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	output, err := session.CombinedOutput("")
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}

	go func() {
		for _, cmd := range commands {
			stdin.Write([]byte(cmd + "\n"))
			time.Sleep(100 * time.Millisecond)
		}
		stdin.Close()
	}()

	return string(output), nil
}

func (s *PerElementStrategy) buildJuniperCommands(elements []ConfigElement) []string {
	commands := []string{"configure"}

	for _, element := range elements {
		switch element.Operation {
		case "set":
			commands = append(commands, fmt.Sprintf("set %s %s", element.Path, element.Value))
		case "delete":
			commands = append(commands, fmt.Sprintf("delete %s", element.Path))
		case "edit":
			commands = append(commands, fmt.Sprintf("edit %s", element.Path))
			commands = append(commands, fmt.Sprintf("set %s", element.Value))
			commands = append(commands, "up")
		}
	}

	commands = append(commands, "commit and-quit")
	return commands
}

func (s *PerElementStrategy) Deploy(device *Device, config string) error {
	if device.Vendor != "juniper" {
		return fmt.Errorf("per-element update only supported for Juniper devices")
	}

	elements, err := s.parseConfigToElements(config, device)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	if len(elements) == 0 {
		return fmt.Errorf("no configuration elements to deploy")
	}

	client, err := s.connectSSH(device)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("Deploying %d configuration elements to %s (%s)", len(elements), device.Hostname, device.ManagementIP)

	snapshotID, err := s.createSnapshot(client)
	if err != nil {
		log.Printf("Warning: failed to create snapshot: %v", err)
	} else {
		log.Printf("Created configuration snapshot: %s", snapshotID)
	}

	for i, element := range elements {
		log.Printf("Applying element %d/%d: %s %s", i+1, len(elements), element.Operation, element.Path)

		if err := s.applyElement(client, element); err != nil {
			log.Printf("Failed to apply element, attempting rollback...")
			if snapshotID != "" {
				s.rollbackToSnapshot(client, snapshotID)
			}
			return fmt.Errorf("failed to apply element %s: %w", element.Path, err)
		}
	}

	if err := s.commitConfiguration(client); err != nil {
		log.Printf("Commit failed, rolling back...")
		if snapshotID != "" {
			s.rollbackToSnapshot(client, snapshotID)
		}
		return fmt.Errorf("failed to commit configuration: %w", err)
	}

	log.Printf("Successfully deployed configuration to %s", device.Hostname)
	return nil
}

func (s *PerElementStrategy) parseConfigToElements(config string, device *Device) ([]ConfigElement, error) {
	elements := []ConfigElement{}

	for _, iface := range device.Interfaces {
		if iface.Description != "" {
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s", iface.Name),
				Operation:   "set",
				Value:       fmt.Sprintf("description \"%s\"", iface.Description),
				Description: fmt.Sprintf("Set description for %s", iface.Name),
			})
		}

		if iface.IPAddress != "" {
			cidrBits := getMaskBits(iface.SubnetMask)
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s unit 0 family inet", iface.Name),
				Operation:   "set",
				Value:       fmt.Sprintf("address %s/%s", iface.IPAddress, cidrBits),
				Description: fmt.Sprintf("Set IP address for %s", iface.Name),
			})
		}

		if iface.MTU > 0 {
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s", iface.Name),
				Operation:   "set",
				Value:       fmt.Sprintf("mtu %d", iface.MTU),
				Description: fmt.Sprintf("Set MTU for %s", iface.Name),
			})
		}

		if !iface.Enabled {
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s", iface.Name),
				Operation:   "set",
				Value:       "disable",
				Description: fmt.Sprintf("Disable %s", iface.Name),
			})
		}
	}

	if device.Services.NTP.Enabled {
		for _, server := range device.Services.NTP.Servers {
			elements = append(elements, ConfigElement{
				Type:        "service",
				Path:        "system ntp",
				Operation:   "set",
				Value:       fmt.Sprintf("server %s", server),
				Description: fmt.Sprintf("Add NTP server %s", server),
			})
		}
	}

	if device.Services.SNMP.Enabled {
		elements = append(elements, ConfigElement{
			Type:        "service",
			Path:        fmt.Sprintf("snmp community %s", device.Services.SNMP.Community),
			Operation:   "set",
			Value:       "authorization read-only",
			Description: "Configure SNMP community",
		})

		elements = append(elements, ConfigElement{
			Type:        "service",
			Path:        "snmp",
			Operation:   "set",
			Value:       fmt.Sprintf("location \"%s\"", device.Services.SNMP.Location),
			Description: "Set SNMP location",
		})

		elements = append(elements, ConfigElement{
			Type:        "service",
			Path:        "snmp",
			Operation:   "set",
			Value:       fmt.Sprintf("contact \"%s\"", device.Services.SNMP.Contact),
			Description: "Set SNMP contact",
		})
	}

	for _, vlan := range device.VLANs {
		elements = append(elements, ConfigElement{
			Type:        "vlan",
			Path:        fmt.Sprintf("vlans %s", vlan.Name),
			Operation:   "set",
			Value:       fmt.Sprintf("vlan-id %d", vlan.ID),
			Description: fmt.Sprintf("Configure VLAN %s", vlan.Name),
		})

		if vlan.Description != "" {
			elements = append(elements, ConfigElement{
				Type:        "vlan",
				Path:        fmt.Sprintf("vlans %s", vlan.Name),
				Operation:   "set",
				Value:       fmt.Sprintf("description \"%s\"", vlan.Description),
				Description: fmt.Sprintf("Set VLAN %s description", vlan.Name),
			})
		}
	}

	if device.Routing != nil {
		for _, route := range device.Routing.StaticRoutes {
			elements = append(elements, ConfigElement{
				Type:        "routing",
				Path:        "routing-options static",
				Operation:   "set",
				Value:       fmt.Sprintf("route %s next-hop %s", route.Destination, route.NextHop),
				Description: fmt.Sprintf("Add static route to %s", route.Destination),
			})
		}
	}

	return elements, nil
}

func (s *PerElementStrategy) applyElement(client *ssh.Client, element ConfigElement) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var command string
	switch element.Operation {
	case "set":
		command = fmt.Sprintf("configure; set %s %s; commit check; exit", element.Path, element.Value)
	case "delete":
		command = fmt.Sprintf("configure; delete %s; commit check; exit", element.Path)
	default:
		return fmt.Errorf("unsupported operation: %s", element.Operation)
	}

	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}

	if strings.Contains(string(output), "error") || strings.Contains(string(output), "invalid") {
		return fmt.Errorf("configuration check failed: %s", string(output))
	}

	return nil
}

func (s *PerElementStrategy) commitConfiguration(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	command := "configure; commit and-quit"
	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("commit failed: %w, output: %s", err, string(output))
	}

	if strings.Contains(string(output), "error") {
		return fmt.Errorf("commit error: %s", string(output))
	}

	log.Printf("Configuration committed successfully")
	return nil
}

func (s *PerElementStrategy) createSnapshot(client *ssh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	snapshotID := fmt.Sprintf("snapshot-%d", time.Now().Unix())
	command := fmt.Sprintf("request system snapshot slice alternate media internal")

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("snapshot failed: %w", err)
	}

	log.Printf("Snapshot output: %s", string(output))
	return snapshotID, nil
}

func (s *PerElementStrategy) rollbackToSnapshot(client *ssh.Client, snapshotID string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	command := "configure; rollback; commit and-quit"
	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Printf("Rolled back configuration: %s", string(output))
	return nil
}

func (s *PerElementStrategy) Rollback(device *Device, backupConfig string) error {
	client, err := s.connectSSH(device)
	if err != nil {
		return err
	}
	defer client.Close()

	return s.rollbackToSnapshot(client, "")
}

type ElementUpdateResult struct {
	Element   ConfigElement
	Success   bool
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

type PerElementDeployer struct {
	strategy *PerElementStrategy
}

func NewPerElementDeployer(strategy *PerElementStrategy) *PerElementDeployer {
	return &PerElementDeployer{
		strategy: strategy,
	}
}

func (d *PerElementDeployer) DeployElements(device *Device, elements []ConfigElement) ([]ElementUpdateResult, error) {
	results := make([]ElementUpdateResult, 0, len(elements))

	client, err := d.strategy.connectSSH(device)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	snapshotID, err := d.strategy.createSnapshot(client)
	if err != nil {
		log.Printf("Warning: failed to create snapshot: %v", err)
	}

	for _, element := range elements {
		startTime := time.Now()
		result := ElementUpdateResult{
			Element:   element,
			Timestamp: startTime,
		}

		err := d.strategy.applyElement(client, element)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Success = false
			result.Error = err
			log.Printf("Failed to apply element %s: %v", element.Path, err)

			if snapshotID != "" {
				log.Printf("Rolling back due to failure...")
				d.strategy.rollbackToSnapshot(client, snapshotID)
			}

			results = append(results, result)
			return results, fmt.Errorf("element update failed: %w", err)
		}

		result.Success = true
		results = append(results, result)
		log.Printf("Successfully applied element %s", element.Path)
	}

	if err := d.strategy.commitConfiguration(client); err != nil {
		if snapshotID != "" {
			d.strategy.rollbackToSnapshot(client, snapshotID)
		}
		return results, fmt.Errorf("commit failed: %w", err)
	}

	return results, nil
}

func (d *PerElementDeployer) UpdateInterface(device *Device, interfaceName string, updates map[string]interface{}) error {
	elements := []ConfigElement{}

	if desc, ok := updates["description"].(string); ok {
		elements = append(elements, ConfigElement{
			Type:        "interface",
			Path:        fmt.Sprintf("interfaces %s", interfaceName),
			Operation:   "set",
			Value:       fmt.Sprintf("description \"%s\"", desc),
			Description: "Update interface description",
		})
	}

	if mtu, ok := updates["mtu"].(int); ok {
		elements = append(elements, ConfigElement{
			Type:        "interface",
			Path:        fmt.Sprintf("interfaces %s", interfaceName),
			Operation:   "set",
			Value:       fmt.Sprintf("mtu %d", mtu),
			Description: "Update interface MTU",
		})
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		if !enabled {
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s", interfaceName),
				Operation:   "set",
				Value:       "disable",
				Description: "Disable interface",
			})
		} else {
			elements = append(elements, ConfigElement{
				Type:        "interface",
				Path:        fmt.Sprintf("interfaces %s disable", interfaceName),
				Operation:   "delete",
				Value:       "",
				Description: "Enable interface",
			})
		}
	}

	if len(elements) == 0 {
		return fmt.Errorf("no valid updates provided")
	}

	_, err := d.DeployElements(device, elements)
	return err
}

func main() {
	model, err := LoadModel("infrastructure.json")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	var juniperDevice *Device
	for i := range model.Devices {
		if model.Devices[i].Vendor == "juniper" {
			juniperDevice = &model.Devices[i]
			break
		}
	}

	if juniperDevice == nil {
		log.Fatalf("No Juniper device found in model")
	}

	strategy := NewPerElementStrategy("admin", "password", 30*time.Second)
	deployer := NewPerElementDeployer(strategy)

	updates := map[string]interface{}{
		"description": "Updated via per-element strategy",
		"mtu":         9000,
	}

	err = deployer.UpdateInterface(juniperDevice, "ge-0/0/0", updates)
	if err != nil {
		log.Fatalf("Failed to update interface: %v", err)
	}

	fmt.Println("Interface updated successfully")
}