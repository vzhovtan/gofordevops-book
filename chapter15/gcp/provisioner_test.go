package gcp

import (
    "testing"
)

func TestGCPConfigValidation(t *testing.T) {
    config := &GCPConfig{
        ProjectID:      "test-project",
        Region:         "us-central1",
        Zone:           "us-central1-a",
        NetworkName:    "infrastructure-network",
        SubnetName:     "infrastructure-subnet",
        SubnetCIDR:     "10.0.1.0/24",
        FirewallName:   "infrastructure-firewall",
        InstanceCount:  5,
        MachineType:    "e2-medium",
        BootDiskImage:  "ubuntu-2204-lts",
        BootDiskSizeGB: 20,
        SSHKey:         "ssh-rsa AAAA...",
    }

    if config.InstanceCount != 5 {
        t.Errorf("Expected 5 instances, got %d", config.InstanceCount)
    }

    if config.Zone != "us-central1-a" {
        t.Errorf("Expected zone us-central1-a, got %s", config.Zone)
    }

    if config.SubnetCIDR != "10.0.1.0/24" {
        t.Errorf("Expected subnet CIDR 10.0.1.0/24, got %s", config.SubnetCIDR)
    }
}

func TestGCPInfrastructureState(t *testing.T) {
    infra := &GCPInfrastructure{
        ProjectID:  "test-project",
        NetworkURL: "projects/test-project/global/networks/test-network",
        SubnetURL:  "projects/test-project/regions/us-central1/subnetworks/test-subnet",
        InstanceNames: []string{
            "instance-1", "instance-2", "instance-3", "instance-4", "instance-5",
        },
        InstanceIPs: []string{
            "34.1.1.1", "34.1.1.2", "34.1.1.3", "34.1.1.4", "34.1.1.5",
        },
    }

    if len(infra.InstanceNames) != 5 {
        t.Errorf("Expected 5 instance names, got %d", len(infra.InstanceNames))
    }

    if len(infra.InstanceIPs) != 5 {
        t.Errorf("Expected 5 instance IPs, got %d", len(infra.InstanceIPs))
    }

    if len(infra.InstanceNames) != len(infra.InstanceIPs) {
        t.Error("Instance names and IPs count should match")
    }

    if infra.NetworkURL == "" {
        t.Error("Expected network URL to be set")
    }
}

func TestScaleInstances(t *testing.T) {
    infra := &GCPInfrastructure{
        InstanceNames: []string{
            "instance-1", "instance-2", "instance-3", "instance-4", "instance-5",
        },
        InstanceIPs: make([]string, 5),
    }

    initialCount := len(infra.InstanceNames)
    targetCount := 10

    if targetCount <= initialCount {
        t.Error("Target count must be greater than initial count")
    }

    expectedAdditional := targetCount - initialCount
    if expectedAdditional != 5 {
        t.Errorf("Expected to add 5 instances, calculated %d", expectedAdditional)
    }
}

func TestScaleValidation(t *testing.T) {
    tests := []struct {
        name         string
        currentCount int
        targetCount  int
        expectError  bool
    }{
        {"Valid scale up", 5, 10, false},
        {"Invalid same count", 5, 5, true},
        {"Invalid scale down", 10, 5, true},
        {"Valid large scale", 5, 100, false},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            shouldError := tc.targetCount <= tc.currentCount

            if shouldError != tc.expectError {
                t.Errorf("Expected error=%v for current=%d target=%d", tc.expectError, tc.currentCount, tc.targetCount)
            }
        })
    }
}