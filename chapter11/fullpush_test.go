package infra

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type MockDeploymentStrategy struct {
	deployError    error
	rollbackError  error
	deployCalled   bool
	rollbackCalled bool
	deployedConfig string
	backupConfig   string
}

func (m *MockDeploymentStrategy) Deploy(device *Device, config string) error {
	m.deployCalled = true
	m.deployedConfig = config
	return m.deployError
}

func (m *MockDeploymentStrategy) Rollback(device *Device, backupConfig string) error {
	m.rollbackCalled = true
	m.backupConfig = backupConfig
	return m.rollbackError
}

func createTestDeviceForDeployment(vendor string) *Device {
	return &Device{
		ID:           "test-deploy-01",
		Hostname:     "test-device",
		DeviceType:   "router",
		Vendor:       vendor,
		Model:        "TestModel",
		ManagementIP: "192.168.1.1",
		Location: Location{
			Datacenter: "dc-test",
			Rack:       "R01",
			Position:   "U10",
		},
	}
}

func TestNewFullReplaceStrategy(t *testing.T) {
	strategy := NewFullReplaceStrategy("admin", "password", 30*time.Second)
	
	if strategy == nil {
		t.Fatal("Expected non-nil strategy")
	}
	
	if strategy.sshConfig == nil {
		t.Error("Expected SSH config to be initialized")
	}
	
	if strategy.sshConfig.User != "admin" {
		t.Errorf("Expected user 'admin', got %s", strategy.sshConfig.User)
	}
	
	if strategy.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", strategy.timeout)
	}
}

func TestNewConfigDeployer(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{}
	deployer := NewConfigDeployer(mockStrategy)
	
	if deployer == nil {
		t.Fatal("Expected non-nil deployer")
	}
	
	if deployer.strategy == nil {
		t.Error("Expected strategy to be set")
	}
}

func TestDeployToDeviceSuccess(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{
		deployError: nil,
	}
	deployer := NewConfigDeployer(mockStrategy)
	
	device := createTestDeviceForDeployment("cisco")
	config := "hostname test-device\n"
	
	result := deployer.DeployToDevice(device, config)
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if !result.Success {
		t.Error("Expected successful deployment")
	}
	
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	
	if result.DeviceID != device.ID {
		t.Errorf("Expected device ID %s, got %s", device.ID, result.DeviceID)
	}
	
	if !mockStrategy.deployCalled {
		t.Error("Expected Deploy to be called")
	}
	
	if mockStrategy.deployedConfig != config {
		t.Errorf("Expected config %s, got %s", config, mockStrategy.deployedConfig)
	}
}

func TestDeployToDeviceFailure(t *testing.T) {
	expectedError := errors.New("deployment failed")
	mockStrategy := &MockDeploymentStrategy{
		deployError: expectedError,
	}
	deployer := NewConfigDeployer(mockStrategy)
	
	device := createTestDeviceForDeployment("cisco")
	config := "hostname test-device\n"
	
	result := deployer.DeployToDevice(device, config)
	
	if result.Success {
		t.Error("Expected failed deployment")
	}
	
	if result.Error == nil {
		t.Error("Expected error to be set")
	}
	
	if !mockStrategy.deployCalled {
		t.Error("Expected Deploy to be called")
	}
}

func TestDeployToDeviceTimestamp(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{}
	deployer := NewConfigDeployer(mockStrategy)
	
	device := createTestDeviceForDeployment("cisco")
	config := "hostname test-device\n"
	
	beforeDeploy := time.Now()
	result := deployer.DeployToDevice(device, config)
	afterDeploy := time.Now()
	
	if result.Timestamp.Before(beforeDeploy) {
		t.Error("Timestamp should be after deployment started")
	}
	
	if result.Timestamp.After(afterDeploy) {
		t.Error("Timestamp should be before deployment completed")
	}
	
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestDeployToMultipleDevicesSuccess(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{}
	deployer := NewConfigDeployer(mockStrategy)
	
	devices := []*Device{
		createTestDeviceForDeployment("cisco"),
		createTestDeviceForDeployment("cisco"),
	}
	devices[0].ID = "device-01"
	devices[1].ID = "device-02"
	
	configs := map[string]string{
		"device-01": "hostname device-01\n",
		"device-02": "hostname device-02\n",
	}
	
	results := deployer.DeployToMultipleDevices(devices, configs)
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	for _, result := range results {
		if !result.Success {
			t.Errorf("Expected success for device %s", result.DeviceID)
		}
	}
}

func TestDeployToMultipleDevicesStopsOnFailure(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{
		deployError: errors.New("deployment failed"),
	}
	deployer := NewConfigDeployer(mockStrategy)
	
	devices := []*Device{
		createTestDeviceForDeployment("cisco"),
		createTestDeviceForDeployment("cisco"),
		createTestDeviceForDeployment("cisco"),
	}
	devices[0].ID = "device-01"
	devices[1].ID = "device-02"
	devices[2].ID = "device-03"
	
	configs := map[string]string{
		"device-01": "config-01",
		"device-02": "config-02",
		"device-03": "config-03",
	}
	
	results := deployer.DeployToMultipleDevices(devices, configs)
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result (stopped on first failure), got %d", len(results))
	}
	
	if results[0].Success {
		t.Error("Expected first deployment to fail")
	}
}

func TestDeployToMultipleDevicesMissingConfig(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{}
	deployer := NewConfigDeployer(mockStrategy)
	
	devices := []*Device{
		createTestDeviceForDeployment("cisco"),
		createTestDeviceForDeployment("cisco"),
	}
	devices[0].ID = "device-01"
	devices[1].ID = "device-02"
	
	configs := map[string]string{
		"device-01": "config-01",
	}
	
	results := deployer.DeployToMultipleDevices(devices, configs)
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result (device-02 skipped), got %d", len(results))
	}
	
	if results[0].DeviceID != "device-01" {
		t.Error("Expected only device-01 to be deployed")
	}
}

func TestNormalizeConfig(t *testing.T) {
	config := `
! Comment line
hostname test-router

interface GigabitEthernet0/0/0
 description Test interface
 ip address 192.168.1.1 255.255.255.0

!
! Another comment
end
`
	
	normalized := normalizeConfig(config)
	
	for _, line := range normalized {
		if strings.HasPrefix(line, "!") {
			t.Error("Normalized config should not contain comment lines")
		}
		
		if line == "" {
			t.Error("Normalized config should not contain empty lines")
		}
	}
	
	if !containsLine(normalized, "hostname test-router") {
		t.Error("Normalized config should contain hostname line")
	}
}

func TestIsCriticalLine(t *testing.T) {
	tests := []struct {
		line     string
		critical bool
	}{
		{"hostname test-router", true},
		{"interface GigabitEthernet0/0/0", true},
		{"ip address 192.168.1.1 255.255.255.0", true},
		{"router ospf 100", true},
		{"network 192.168.1.0 0.0.0.255 area 0", true},
		{"description Test interface", false},
		{"mtu 1500", false},
		{"shutdown", false},
		{"end", false},
	}
	
	for _, tc := range tests {
		result := isCriticalLine(tc.line)
		if result != tc.critical {
			t.Errorf("isCriticalLine(%q) = %v, expected %v", tc.line, result, tc.critical)
		}
	}
}

func TestContainsLine(t *testing.T) {
	lines := []string{
		"hostname test-router",
		"interface GigabitEthernet0/0/0",
		"ip address 192.168.1.1 255.255.255.0",
	}
	
	tests := []struct {
		target   string
		expected bool
	}{
		{"hostname test-router", true},
		{"test-router", true},
		{"GigabitEthernet0/0/0", true},
		{"192.168.1.1", true},
		{"nonexistent", false},
		{"hostname other-router", false},
	}
	
	for _, tc := range tests {
		result := containsLine(lines, tc.target)
		if result != tc.expected {
			t.Errorf("containsLine(..., %q) = %v, expected %v", tc.target, result, tc.expected)
		}
	}
}

func TestConfigBackup(t *testing.T) {
	backup := &ConfigBackup{
		DeviceID:  "test-device",
		Timestamp: time.Now(),
		Config:    "hostname test-router\n",
	}
	
	if backup.DeviceID != "test-device" {
		t.Errorf("Expected device ID test-device, got %s", backup.DeviceID)
	}
	
	if backup.Config == "" {
		t.Error("Expected non-empty config")
	}
	
	if backup.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestDeploymentResultStructure(t *testing.T) {
	result := &DeploymentResult{
		DeviceID:  "test-device",
		Success:   true,
		Error:     nil,
		Duration:  5 * time.Second,
		Timestamp: time.Now(),
		Backup: &ConfigBackup{
			DeviceID:  "test-device",
			Timestamp: time.Now(),
			Config:    "backup config",
		},
	}
	
	if result.DeviceID != "test-device" {
		t.Errorf("Expected device ID test-device, got %s", result.DeviceID)
	}
	
	if !result.Success {
		t.Error("Expected success to be true")
	}
	
	if result.Duration != 5*time.Second {
		t.Errorf("Expected duration 5s, got %v", result.Duration)
	}
	
	if result.Backup == nil {
		t.Error("Expected backup to be set")
	}
}

func TestConfigMatchesWithCriticalLines(t *testing.T) {
	strategy := &FullReplaceStrategy{}
	
	current := `
hostname test-router
interface GigabitEthernet0/0/0
 description Test interface
 ip address 192.168.1.1 255.255.255.0
router ospf 100
 network 192.168.1.0 0.0.0.255 area 0
`
	
	expected := `
hostname test-router
interface GigabitEthernet0/0/0
 ip address 192.168.1.1 255.255.255.0
router ospf 100
`
	
	if !strategy.configMatches(current, expected) {
		t.Error("Expected configs to match on critical lines")
	}
}

func TestConfigMatchesFailsWithoutCriticalLines(t *testing.T) {
	strategy := &FullReplaceStrategy{}
	
	current := `
description Some interface
mtu 1500
shutdown
`
	
	expected := `
hostname test-router
interface GigabitEthernet0/0/0
`
	
	if strategy.configMatches(current, expected) {
		t.Error("Expected configs not to match without critical lines")
	}
}

func TestMockStrategyRollback(t *testing.T) {
	mockStrategy := &MockDeploymentStrategy{}
	
	device := createTestDeviceForDeployment("cisco")
	backupConfig := "hostname backup-config\n"
	
	err := mockStrategy.Rollback(device, backupConfig)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if !mockStrategy.rollbackCalled {
		t.Error("Expected Rollback to be called")
	}
	
	if mockStrategy.backupConfig != backupConfig {
		t.Errorf("Expected backup config %s, got %s", backupConfig, mockStrategy.backupConfig)
	}
}

func TestMockStrategyRollbackWithError(t *testing.T) {
	expectedError := errors.New("rollback failed")
	mockStrategy := &MockDeploymentStrategy{
		rollbackError: expectedError,
	}
	
	device := createTestDeviceForDeployment("cisco")
	
	err := mockStrategy.Rollback(device, "config")
	
	if err == nil {
		t.Error("Expected error from rollback")
	}
	
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestDeploymentResultWithError(t *testing.T) {
	deployError := errors.New("test deployment error")
	mockStrategy := &MockDeploymentStrategy{
		deployError: deployError,
	}
	deployer := NewConfigDeployer(mockStrategy)
	
	device := createTestDeviceForDeployment("cisco")
	config := "hostname test\n"
	
	result := deployer.DeployToDevice(device, config)
	
	if result.Success {
		t.Error("Expected deployment to fail")
	}
	
	if result.Error == nil {
		t.Error("Expected error to be set in result")
	}
	
	if result.Error != deployError {
		t.Errorf("Expected error %v, got %v", deployError, result.Error)
	}
}

func TestMultipleDeploymentsIndependence(t *testing.T) {
	callCount := 0
	mockStrategy := &MockDeploymentStrategy{}
	
	originalDeploy := mockStrategy.Deploy
	mockStrategy.Deploy = func(device *Device, config string) error {
		callCount++
		if callCount == 1 {
			return errors.New("first deployment fails")
		}
		return originalDeploy(device, config)
	}
	
	deployer := NewConfigDeployer(mockStrategy)
	
	devices := []*Device{
		createTestDeviceForDeployment("cisco"),
		createTestDeviceForDeployment("cisco"),
	}
	devices[0].ID = "device-01"
	devices[1].ID = "device-02"
	
	configs := map[string]string{
		"device-01": "config-01",
		"device-02": "config-02",
	}
	
	results := deployer.DeployToMultipleDevices(devices, configs)
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	
	if results[0].Success {
		t.Error("Expected first deployment to fail")
	}
	
	if callCount != 1 {
		t.Errorf("Expected Deploy to be called once, was called %d times", callCount)
	}
}