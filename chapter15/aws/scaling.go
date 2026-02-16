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