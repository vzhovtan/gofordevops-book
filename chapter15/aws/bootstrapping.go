package aws

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InfrastructureConfig struct {
    VPCCidr        string
    SubnetCidr     string
    InstanceCount  int
    InstanceType   types.InstanceType
    AMI            string
    KeyName        string
    Region         string
}

type Infrastructure struct {
    VpcID              *string
    SubnetID           *string
    InternetGatewayID  *string
    RouteTableID       *string
    SecurityGroupID    *string
    InstanceIDs        []string
}

type AWSProvisioner struct {
    ec2Client *ec2.Client
    config    *InfrastructureConfig
}

func NewAWSProvisioner(cfg *InfrastructureConfig) (*AWSProvisioner, error) {
    ctx := context.Background()
    
    awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    return &AWSProvisioner{
        ec2Client: ec2.NewFromConfig(awsCfg),
        config:    cfg,
    }, nil
}

func (p *AWSProvisioner) ProvisionInfrastructure(ctx context.Context) (*Infrastructure, error) {
    infra := &Infrastructure{}

    vpcID, err := p.createVPC(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create VPC: %w", err)
    }
    infra.VpcID = vpcID

    igwID, err := p.createInternetGateway(ctx, vpcID)
    if err != nil {
        return nil, fmt.Errorf("failed to create internet gateway: %w", err)
    }
    infra.InternetGatewayID = igwID

    subnetID, err := p.createSubnet(ctx, vpcID)
    if err != nil {
        return nil, fmt.Errorf("failed to create subnet: %w", err)
    }
    infra.SubnetID = subnetID

    rtID, err := p.createRouteTable(ctx, vpcID, igwID)
    if err != nil {
        return nil, fmt.Errorf("failed to create route table: %w", err)
    }
    infra.RouteTableID = rtID

    if err := p.associateRouteTable(ctx, rtID, subnetID); err != nil {
        return nil, fmt.Errorf("failed to associate route table: %w", err)
    }

    sgID, err := p.createSecurityGroup(ctx, vpcID)
    if err != nil {
        return nil, fmt.Errorf("failed to create security group: %w", err)
    }
    infra.SecurityGroupID = sgID

    instanceIDs, err := p.launchInstances(ctx, subnetID, sgID, p.config.InstanceCount)
    if err != nil {
        return nil, fmt.Errorf("failed to launch instances: %w", err)
    }
    infra.InstanceIDs = instanceIDs

    return infra, nil
}

func (p *AWSProvisioner) createVPC(ctx context.Context) (*string, error) {
    result, err := p.ec2Client.CreateVpc(ctx, &ec2.CreateVpcInput{
        CidrBlock: aws.String(p.config.VPCCidr),
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeVpc,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("infrastructure-vpc")},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    // Enable DNS hostnames
    _, err = p.ec2Client.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
        VpcId:              result.Vpc.VpcId,
        EnableDnsHostnames: &types.AttributeBooleanValue{Value: aws.Bool(true)},
    })
    if err != nil {
        return nil, err
    }

    return result.Vpc.VpcId, nil
}

func (p *AWSProvisioner) createInternetGateway(ctx context.Context, vpcID *string) (*string, error) {
    result, err := p.ec2Client.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeInternetGateway,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("infrastructure-igw")},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    _, err = p.ec2Client.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
        InternetGatewayId: result.InternetGateway.InternetGatewayId,
        VpcId:             vpcID,
    })
    if err != nil {
        return nil, err
    }

    return result.InternetGateway.InternetGatewayId, nil
}

func (p *AWSProvisioner) createSubnet(ctx context.Context, vpcID *string) (*string, error) {
    result, err := p.ec2Client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
        VpcId:     vpcID,
        CidrBlock: aws.String(p.config.SubnetCidr),
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeSubnet,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("infrastructure-subnet")},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    return result.Subnet.SubnetId, nil
}

func (p *AWSProvisioner) createRouteTable(ctx context.Context, vpcID, igwID *string) (*string, error) {
    result, err := p.ec2Client.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
        VpcId: vpcID,
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeRouteTable,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("infrastructure-rt")},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    _, err = p.ec2Client.CreateRoute(ctx, &ec2.CreateRouteInput{
        RouteTableId:         result.RouteTable.RouteTableId,
        DestinationCidrBlock: aws.String("0.0.0.0/0"),
        GatewayId:            igwID,
    })
    if err != nil {
        return nil, err
    }

    return result.RouteTable.RouteTableId, nil
}

func (p *AWSProvisioner) associateRouteTable(ctx context.Context, rtID, subnetID *string) error {
    _, err := p.ec2Client.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
        RouteTableId: rtID,
        SubnetId:     subnetID,
    })
    return err
}

func (p *AWSProvisioner) createSecurityGroup(ctx context.Context, vpcID *string) (*string, error) {
    result, err := p.ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
        GroupName:   aws.String("infrastructure-sg"),
        Description: aws.String("Security group for infrastructure instances"),
        VpcId:       vpcID,
    })
    if err != nil {
        return nil, err
    }

    _, err = p.ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
        GroupId: result.GroupId,
        IpPermissions: []types.IpPermission{
            {
                IpProtocol: aws.String("tcp"),
                FromPort:   aws.Int32(22),
                ToPort:     aws.Int32(22),
                IpRanges:   []types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
            },
        },
    })
    if err != nil {
        return nil, err
    }

    return result.GroupId, nil
}

func (p *AWSProvisioner) launchInstances(ctx context.Context, subnetID, sgID *string, count int) ([]string, error) {
    result, err := p.ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
        ImageId:      aws.String(p.config.AMI),
        InstanceType: p.config.InstanceType,
        MinCount:     aws.Int32(int32(count)),
        MaxCount:     aws.Int32(int32(count)),
        KeyName:      aws.String(p.config.KeyName),
        NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
            {
                AssociatePublicIpAddress: aws.Bool(true),
                DeviceIndex:              aws.Int32(0),
                SubnetId:                 subnetID,
                Groups:                   []string{*sgID},
            },
        },
        TagSpecifications: []types.TagSpecification{
            {
                ResourceType: types.ResourceTypeInstance,
                Tags: []types.Tag{
                    {Key: aws.String("Name"), Value: aws.String("infrastructure-instance")},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    instanceIDs := make([]string, len(result.Instances))
    for i, instance := range result.Instances {
        instanceIDs[i] = *instance.InstanceId
    }

    return instanceIDs, nil
}

func (p *AWSProvisioner) TeardownInfrastructure(ctx context.Context, infra *Infrastructure) error {
    if len(infra.InstanceIDs) > 0 {
        if err := p.terminateInstances(ctx, infra.InstanceIDs); err != nil {
            return fmt.Errorf("failed to terminate instances: %w", err)
        }
    }

    if infra.SecurityGroupID != nil {
        if err := p.deleteSecurityGroup(ctx, infra.SecurityGroupID); err != nil {
            log.Printf("Warning: failed to delete security group: %v", err)
        }
    }

    if infra.InternetGatewayID != nil {
        if err := p.deleteInternetGateway(ctx, infra.InternetGatewayID, infra.VpcID); err != nil {
            log.Printf("Warning: failed to delete internet gateway: %v", err)
        }
    }

    if infra.SubnetID != nil {
        if err := p.deleteSubnet(ctx, infra.SubnetID); err != nil {
            log.Printf("Warning: failed to delete subnet: %v", err)
        }
    }

    if infra.RouteTableID != nil {
        if err := p.deleteRouteTable(ctx, infra.RouteTableID); err != nil {
            log.Printf("Warning: failed to delete route table: %v", err)
        }
    }

    if infra.VpcID != nil {
        if err := p.deleteVPC(ctx, infra.VpcID); err != nil {
            return fmt.Errorf("failed to delete VPC: %w", err)
        }
    }

    return nil
}

func (p *AWSProvisioner) terminateInstances(ctx context.Context, instanceIDs []string) error {
    _, err := p.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
        InstanceIds: instanceIDs,
    })
    if err != nil {
        return err
    }

    waiter := ec2.NewInstanceTerminatedWaiter(p.ec2Client)
    return waiter.Wait(ctx, &ec2.DescribeInstancesInput{
        InstanceIds: instanceIDs,
    }, 5*time.Minute)
}

func (p *AWSProvisioner) deleteSecurityGroup(ctx context.Context, sgID *string) error {
    _, err := p.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
        GroupId: sgID,
    })
    return err
}

func (p *AWSProvisioner) deleteInternetGateway(ctx context.Context, igwID, vpcID *string) error {
    _, err := p.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
        InternetGatewayId: igwID,
        VpcId:             vpcID,
    })
    if err != nil {
        return err
    }

    _, err = p.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
        InternetGatewayId: igwID,
    })
    return err
}

func (p *AWSProvisioner) deleteSubnet(ctx context.Context, subnetID *string) error {
    _, err := p.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
        SubnetId: subnetID,
    })
    return err
}

func (p *AWSProvisioner) deleteRouteTable(ctx context.Context, rtID *string) error {
    _, err := p.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
        RouteTableId: rtID,
    })
    return err
}

func (p *AWSProvisioner) deleteVPC(ctx context.Context, vpcID *string) error {
    _, err := p.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
        VpcId: vpcID,
    })
    return err
}

func (p *AWSProvisioner) ScaleInstances(ctx context.Context, infra *Infrastructure, targetCount int) error {
    currentCount := len(infra.InstanceIDs)
    
    if targetCount <= currentCount {
        return fmt.Errorf("target count %d must be greater than current count %d", targetCount, currentCount)
    }

    additionalCount := targetCount - currentCount
    
    newInstanceIDs, err := p.launchInstances(ctx, infra.SubnetID, infra.SecurityGroupID, additionalCount)
    if err != nil {
        return fmt.Errorf("failed to launch additional instances: %w", err)
    }

    infra.InstanceIDs = append(infra.InstanceIDs, newInstanceIDs...)
    
    return nil
}

func (p *AWSProvisioner) WaitForInstancesRunning(ctx context.Context, instanceIDs []string) error {
    waiter := ec2.NewInstanceRunningWaiter(p.ec2Client)
    return waiter.Wait(ctx, &ec2.DescribeInstancesInput{
        InstanceIds: instanceIDs,
    }, 5*time.Minute)
}