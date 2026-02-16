package main

import (
    "testing"
)

func TestAzureConfigValidation(t *testing.T) {
    config := &AzureConfig{
        SubscriptionID:    "test-subscription-id",
        Location:          "eastus",
        ResourceGroupName: "infrastructure-rg",
        VNetName:          "infrastructure-vnet",
        VNetCIDR:          "10.0.0.0/16",
        SubnetName:        "infrastructure-subnet",
        SubnetCIDR:        "10.0.1.0/24",
        NSGName:           "infrastructure-nsg",
        VMCount:           5,
        VMSize:            "Standard_B2s",
        AdminUsername:     "azureuser",
        SSHPublicKey:      "ssh-rsa AAAA...",
    }

    if config.VMCount != 5 {
        t.Errorf("Expected 5 VMs, got %d", config.VMCount)
    }

    if config.Location != "eastus" {
        t.Errorf("Expected location eastus, got %s", config.Location)
    }

    if config.VNetCIDR != "10.0.0.0/16" {
        t.Errorf("Expected VNet CIDR 10.0.0.0/16, got %s", config.VNetCIDR)
    }
}

func TestAzureInfrastructureState(t *testing.T) {
    infra := &AzureInfrastructure{
        ResourceGroupName: "test-rg",
        VNetID:            "/subscriptions/.../virtualNetworks/test-vnet",
        SubnetID:          "/subscriptions/.../subnets/test-subnet",
        NSGID:             "/subscriptions/.../networkSecurityGroups/test-nsg",
        VMIDs: []string{
            "/subscriptions/.../virtualMachines/vm-1",
            "/subscriptions/.../virtualMachines/vm-2",
            "/subscriptions/.../virtualMachines/vm-3",
            "/subscriptions/.../virtualMachines/vm-4",
            "/subscriptions/.../virtualMachines/vm-5",
        },
        PublicIPIDs: []string{
            "/subscriptions/.../publicIPAddresses/vm-1-pip",
            "/subscriptions/.../publicIPAddresses/vm-2-pip",
            "/subscriptions/.../publicIPAddresses/vm-3-pip",
            "/subscriptions/.../publicIPAddresses/vm-4-pip",
            "/subscriptions/.../publicIPAddresses/vm-5-pip",
        },
        NICIDs: []string{
            "/subscriptions/.../networkInterfaces/vm-1-nic",
            "/subscriptions/.../networkInterfaces/vm-2-nic",
            "/subscriptions/.../networkInterfaces/vm-3-nic",
            "/subscriptions/.../networkInterfaces/vm-4-nic",
            "/subscriptions/.../networkInterfaces/vm-5-nic",
        },
    }

    if len(infra.VMIDs) != 5 {
        t.Errorf("Expected 5 VM IDs, got %d", len(infra.VMIDs))
    }

    if len(infra.PublicIPIDs) != 5 {
        t.Errorf("Expected 5 Public IP IDs, got %d", len(infra.PublicIPIDs))
    }

    if len(infra.NICIDs) != 5 {
        t.Errorf("Expected 5 NIC IDs, got %d", len(infra.NICIDs))
    }

    if infra.VNetID == "" {
        t.Error("Expected VNet ID to be set")
    }

    if infra.SubnetID == "" {
        t.Error("Expected Subnet ID to be set")
    }
}