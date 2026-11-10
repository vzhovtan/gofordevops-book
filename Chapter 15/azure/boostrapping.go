package azure

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type AzureConfig struct {
	SubscriptionID    string
	Location          string
	ResourceGroupName string
	VNetName          string
	VNetCIDR          string
	SubnetName        string
	SubnetCIDR        string
	NSGName           string
	VMCount           int
	VMSize            string
	AdminUsername     string
	SSHPublicKey      string
}

type AzureInfrastructure struct {
	ResourceGroupName string
	VNetID            string
	SubnetID          string
	NSGID             string
	VMIDs             []string
	PublicIPIDs       []string
	NICIDs            []string
}

type AzureProvisioner struct {
	cred            azcore.TokenCredential
	config          *AzureConfig
	resourcesClient *armresources.ResourceGroupsClient
	vnetsClient     *armnetwork.VirtualNetworksClient
	subnetsClient   *armnetwork.SubnetsClient
	nsgClient       *armnetwork.SecurityGroupsClient
	publicIPClient  *armnetwork.PublicIPAddressesClient
	nicClient       *armnetwork.InterfacesClient
	vmClient        *armcompute.VirtualMachinesClient
}

func NewAzureProvisioner(cfg *AzureConfig) (*AzureProvisioner, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	resourcesClient, err := armresources.NewResourceGroupsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource groups client: %w", err)
	}

	vnetsClient, err := armnetwork.NewVirtualNetworksClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual networks client: %w", err)
	}

	subnetsClient, err := armnetwork.NewSubnetsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnets client: %w", err)
	}

	nsgClient, err := armnetwork.NewSecurityGroupsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NSG client: %w", err)
	}

	publicIPClient, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create public IP client: %w", err)
	}

	nicClient, err := armnetwork.NewInterfacesClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NIC client: %w", err)
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %w", err)
	}

	return &AzureProvisioner{
		cred:            cred,
		config:          cfg,
		resourcesClient: resourcesClient,
		vnetsClient:     vnetsClient,
		subnetsClient:   subnetsClient,
		nsgClient:       nsgClient,
		publicIPClient:  publicIPClient,
		nicClient:       nicClient,
		vmClient:        vmClient,
	}, nil
}

func (p *AzureProvisioner) ProvisionInfrastructure(ctx context.Context) (*AzureInfrastructure, error) {
	infra := &AzureInfrastructure{
		ResourceGroupName: p.config.ResourceGroupName,
	}

	if err := p.createResourceGroup(ctx); err != nil {
		return nil, fmt.Errorf("failed to create resource group: %w", err)
	}

	vnetID, err := p.createVirtualNetwork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual network: %w", err)
	}
	infra.VNetID = vnetID

	subnetID, err := p.createSubnet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}
	infra.SubnetID = subnetID

	nsgID, err := p.createNetworkSecurityGroup(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create NSG: %w", err)
	}
	infra.NSGID = nsgID

	if err := p.associateNSGWithSubnet(ctx, subnetID, nsgID); err != nil {
		return nil, fmt.Errorf("failed to associate NSG: %w", err)
	}

	for i := 0; i < p.config.VMCount; i++ {
		vmName := fmt.Sprintf("vm-%d", i+1)

		publicIPID, err := p.createPublicIP(ctx, fmt.Sprintf("%s-pip", vmName))
		if err != nil {
			return nil, fmt.Errorf("failed to create public IP for %s: %w", vmName, err)
		}
		infra.PublicIPIDs = append(infra.PublicIPIDs, publicIPID)

		nicID, err := p.createNetworkInterface(ctx, fmt.Sprintf("%s-nic", vmName), subnetID, publicIPID, nsgID)
		if err != nil {
			return nil, fmt.Errorf("failed to create NIC for %s: %w", vmName, err)
		}
		infra.NICIDs = append(infra.NICIDs, nicID)

		vmID, err := p.createVirtualMachine(ctx, vmName, nicID)
		if err != nil {
			return nil, fmt.Errorf("failed to create VM %s: %w", vmName, err)
		}
		infra.VMIDs = append(infra.VMIDs, vmID)
	}

	return infra, nil
}

func (p *AzureProvisioner) createResourceGroup(ctx context.Context) error {
	_, err := p.resourcesClient.CreateOrUpdate(ctx, p.config.ResourceGroupName, armresources.ResourceGroup{
		Location: to.Ptr(p.config.Location),
		Tags: map[string]*string{
			"purpose": to.Ptr("infrastructure-automation"),
		},
	}, nil)
	return err
}

func (p *AzureProvisioner) createVirtualNetwork(ctx context.Context) (string, error) {
	pollerResp, err := p.vnetsClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, p.config.VNetName, armnetwork.VirtualNetwork{
		Location: to.Ptr(p.config.Location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{to.Ptr(p.config.VNetCIDR)},
			},
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) createSubnet(ctx context.Context) (string, error) {
	pollerResp, err := p.subnetsClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, p.config.VNetName, p.config.SubnetName, armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr(p.config.SubnetCIDR),
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) createNetworkSecurityGroup(ctx context.Context) (string, error) {
	pollerResp, err := p.nsgClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, p.config.NSGName, armnetwork.SecurityGroup{
		Location: to.Ptr(p.config.Location),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				{
					Name: to.Ptr("allow-ssh"),
					Properties: &armnetwork.SecurityRulePropertiesFormat{
						Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
						SourcePortRange:          to.Ptr("*"),
						DestinationPortRange:     to.Ptr("22"),
						SourceAddressPrefix:      to.Ptr("*"),
						DestinationAddressPrefix: to.Ptr("*"),
						Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
						Priority:                 to.Ptr[int32](100),
						Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) associateNSGWithSubnet(ctx context.Context, subnetID, nsgID string) error {
	pollerResp, err := p.subnetsClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, p.config.VNetName, p.config.SubnetName, armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr(p.config.SubnetCIDR),
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(nsgID),
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	return err
}

func (p *AzureProvisioner) createPublicIP(ctx context.Context, name string) (string, error) {
	pollerResp, err := p.publicIPClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, name, armnetwork.PublicIPAddress{
		Location: to.Ptr(p.config.Location),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) createNetworkInterface(ctx context.Context, name, subnetID, publicIPID, nsgID string) (string, error) {
	pollerResp, err := p.nicClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, name, armnetwork.Interface{
		Location: to.Ptr(p.config.Location),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr("ipconfig1"),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(subnetID),
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.Ptr(publicIPID),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(nsgID),
			},
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) createVirtualMachine(ctx context.Context, name, nicID string) (string, error) {
	pollerResp, err := p.vmClient.BeginCreateOrUpdate(ctx, p.config.ResourceGroupName, name, armcompute.VirtualMachine{
		Location: to.Ptr(p.config.Location),
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(p.config.VMSize)),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Publisher: to.Ptr("Canonical"),
					Offer:     to.Ptr("0001-com-ubuntu-server-jammy"),
					SKU:       to.Ptr("22_04-lts-gen2"),
					Version:   to.Ptr("latest"),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(fmt.Sprintf("%s-osdisk", name)),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					},
				},
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  to.Ptr(name),
				AdminUsername: to.Ptr(p.config.AdminUsername),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", p.config.AdminUsername)),
								KeyData: to.Ptr(p.config.SSHPublicKey),
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.Ptr(nicID),
						Properties: &armcompute.NetworkInterfaceReferenceProperties{
							Primary: to.Ptr(true),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return "", err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return "", err
	}

	return *resp.ID, nil
}

func (p *AzureProvisioner) TeardownInfrastructure(ctx context.Context, infra *AzureInfrastructure) error {
	pollerResp, err := p.resourcesClient.BeginDelete(ctx, infra.ResourceGroupName, nil)
	if err != nil {
		return fmt.Errorf("failed to initiate resource group deletion: %w", err)
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	return nil
}

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
