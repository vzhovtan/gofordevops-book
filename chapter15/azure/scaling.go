func (p *AzureProvisioner) ScaleVirtualMachines(ctx context.Context, infra *AzureInfrastructure, targetCount int) error {
    currentCount := len(infra.VMIDs)
    
    if targetCount <= currentCount {
        return fmt.Errorf("target count %d must be greater than current count %d", targetCount, currentCount)
    }

    additionalCount := targetCount - currentCount

    for i := 0; i < additionalCount; i++ {
        vmIndex := currentCount + i + 1
        vmName := fmt.Sprintf("vm-%d", vmIndex)

        publicIPID, err := p.createPublicIP(ctx, fmt.Sprintf("%s-pip", vmName))
        if err != nil {
            return fmt.Errorf("failed to create public IP for %s: %w", vmName, err)
        }
        infra.PublicIPIDs = append(infra.PublicIPIDs, publicIPID)

        nicID, err := p.createNetworkInterface(ctx, fmt.Sprintf("%s-nic", vmName), infra.SubnetID, publicIPID, infra.NSGID)
        if err != nil {
            return fmt.Errorf("failed to create NIC for %s: %w", vmName, err)
        }
        infra.NICIDs = append(infra.NICIDs, nicID)

        vmID, err := p.createVirtualMachine(ctx, vmName, nicID)
        if err != nil {
            return fmt.Errorf("failed to create VM %s: %w", vmName, err)
        }
        infra.VMIDs = append(infra.VMIDs, vmID)

        log.Printf("Created VM %s (%d/%d)", vmName, i+1, additionalCount)
    }

    return nil
}

func (p *AzureProvisioner) GetVMDetails(ctx context.Context, resourceGroupName, vmName string) (*armcompute.VirtualMachine, error) {
    resp, err := p.vmClient.Get(ctx, resourceGroupName, vmName, nil)
    if err != nil {
        return nil, err
    }
    return &resp.VirtualMachine, nil
}

func (p *AzureProvisioner) ListVMs(ctx context.Context, resourceGroupName string) ([]*armcompute.VirtualMachine, error) {
    pager := p.vmClient.NewListPager(resourceGroupName, nil)
    var vms []*armcompute.VirtualMachine

    for pager.More() {
        page, err := pager.NextPage(ctx)
        if err != nil {
            return nil, err
        }
        vms = append(vms, page.Value...)
    }

    return vms, nil
}