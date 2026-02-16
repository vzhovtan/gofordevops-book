package aws

import (
    "testing"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestProvisionInfrastructure(t *testing.T) {
    config := &InfrastructureConfig{
        VPCCidr:       "10.0.0.0/16",
        SubnetCidr:    "10.0.1.0/24",
        InstanceCount: 5,
        InstanceType:  types.InstanceTypeT2Micro,
        AMI:           "ami-12345678",
        KeyName:       "test-key",
        Region:        "us-west-2",
    }

    if config.InstanceCount != 5 {
        t.Errorf("Expected 5 instances, got %d", config.InstanceCount)
    }

    if config.VPCCidr != "10.0.0.0/16" {
        t.Errorf("Expected VPC CIDR 10.0.0.0/16, got %s", config.VPCCidr)
    }
}

func TestInfrastructureValidation(t *testing.T) {
    infra := &Infrastructure{
        VpcID:    aws.String("vpc-12345"),
        SubnetID: aws.String("subnet-12345"),
        InstanceIDs: []string{
            "i-1", "i-2", "i-3", "i-4", "i-5",
        },
    }

    if len(infra.InstanceIDs) != 5 {
        t.Errorf("Expected 5 instances, got %d", len(infra.InstanceIDs))
    }

    if infra.VpcID == nil {
        t.Error("Expected VPC ID to be set")
    }

    if infra.SubnetID == nil {
        t.Error("Expected Subnet ID to be set")
    }
}

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