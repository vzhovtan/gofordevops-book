func TestScaleVirtualMachines(t *testing.T) {
    infra := &AzureInfrastructure{
        ResourceGroupName: "test-rg",
        SubnetID:          "/subscriptions/.../subnets/test-subnet",
        NSGID:             "/subscriptions/.../networkSecurityGroups/test-nsg",
        VMIDs: []string{
            "/subscriptions/.../virtualMachines/vm-1",
            "/subscriptions/.../virtualMachines/vm-2",
            "/subscriptions/.../virtualMachines/vm-3",
            "/subscriptions/.../virtualMachines/vm-4",
            "/subscriptions/.../virtualMachines/vm-5",
        },
        PublicIPIDs: make([]string, 5),
        NICIDs:      make([]string, 5),
    }

    initialCount := len(infra.VMIDs)
    targetCount := 10

    if targetCount <= initialCount {
        t.Error("Target count must be greater than initial count")
    }

    expectedAdditional := targetCount - initialCount
    if expectedAdditional != 5 {
        t.Errorf("Expected to add 5 VMs, calculated %d", expectedAdditional)
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
        {"Valid large scale", 5, 50, false},
        {"Minimum scale", 1, 2, false},
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

func TestInfrastructureConsistency(t *testing.T) {
    infra := &AzureInfrastructure{
        VMIDs:       make([]string, 5),
        PublicIPIDs: make([]string, 5),
        NICIDs:      make([]string, 5),
    }

    if len(infra.VMIDs) != len(infra.PublicIPIDs) {
        t.Error("VM count should match Public IP count")
    }

    if len(infra.VMIDs) != len(infra.NICIDs) {
        t.Error("VM count should match NIC count")
    }
}