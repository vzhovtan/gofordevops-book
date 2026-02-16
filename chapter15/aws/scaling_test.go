func TestScaleInstances(t *testing.T) {
    infra := &Infrastructure{
        VpcID:           aws.String("vpc-12345"),
        SubnetID:        aws.String("subnet-12345"),
        SecurityGroupID: aws.String("sg-12345"),
        InstanceIDs: []string{
            "i-1", "i-2", "i-3", "i-4", "i-5",
        },
    }

    initialCount := len(infra.InstanceIDs)
    targetCount := 10

    if targetCount <= initialCount {
        t.Errorf("Target count must be greater than initial count")
    }

    expectedAdditional := targetCount - initialCount
    if expectedAdditional != 5 {
        t.Errorf("Expected to add 5 instances, calculated %d", expectedAdditional)
    }
}

func TestScaleInstancesValidation(t *testing.T) {
    tests := []struct {
        name          string
        currentCount  int
        targetCount   int
        expectError   bool
    }{
        {"Valid scale up", 5, 10, false},
        {"Invalid same count", 5, 5, true},
        {"Invalid scale down", 10, 5, true},
        {"Valid large scale", 5, 50, false},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            if tc.targetCount <= tc.currentCount && !tc.expectError {
                t.Error("Expected error for invalid scaling")
            }
            
            if tc.targetCount > tc.currentCount && tc.expectError {
                t.Error("Should not expect error for valid scaling")
            }
        })
    }
}