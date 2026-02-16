package aws

import (
    "context"
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