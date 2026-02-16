package gcp

import (
	"context"
	"fmt"
	"log"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

type GCPConfig struct {
	ProjectID      string
	Region         string
	Zone           string
	NetworkName    string
	SubnetName     string
	SubnetCIDR     string
	FirewallName   string
	InstanceCount  int
	MachineType    string
	BootDiskImage  string
	BootDiskSizeGB int64
	SSHKey         string
}

type GCPInfrastructure struct {
	ProjectID     string
	NetworkURL    string
	SubnetURL     string
	FirewallURL   string
	InstanceNames []string
	InstanceIPs   []string
}

type GCPProvisioner struct {
	config            *GCPConfig
	networksClient    *compute.NetworksClient
	subnetworksClient *compute.SubnetworksClient
	firewallsClient   *compute.FirewallsClient
	instancesClient   *compute.InstancesClient
	addressesClient   *compute.AddressesClient
}

func NewGCPProvisioner(ctx context.Context, cfg *GCPConfig) (*GCPProvisioner, error) {
	networksClient, err := compute.NewNetworksRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create networks client: %w", err)
	}

	subnetworksClient, err := compute.NewSubnetworksRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnetworks client: %w", err)
	}

	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewalls client: %w", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}

	addressesClient, err := compute.NewAddressesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create addresses client: %w", err)
	}

	return &GCPProvisioner{
		config:            cfg,
		networksClient:    networksClient,
		subnetworksClient: subnetworksClient,
		firewallsClient:   firewallsClient,
		instancesClient:   instancesClient,
		addressesClient:   addressesClient,
	}, nil
}

func (p *GCPProvisioner) Close() error {
	if err := p.networksClient.Close(); err != nil {
		return err
	}
	if err := p.subnetworksClient.Close(); err != nil {
		return err
	}
	if err := p.firewallsClient.Close(); err != nil {
		return err
	}
	if err := p.instancesClient.Close(); err != nil {
		return err
	}
	if err := p.addressesClient.Close(); err != nil {
		return err
	}
	return nil
}

func (p *GCPProvisioner) ProvisionInfrastructure(ctx context.Context) (*GCPInfrastructure, error) {
	infra := &GCPInfrastructure{
		ProjectID: p.config.ProjectID,
	}

	networkURL, err := p.createNetwork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}
	infra.NetworkURL = networkURL

	subnetURL, err := p.createSubnet(ctx, networkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}
	infra.SubnetURL = subnetURL

	firewallURL, err := p.createFirewallRule(ctx, networkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall rule: %w", err)
	}
	infra.FirewallURL = firewallURL

	for i := 0; i < p.config.InstanceCount; i++ {
		instanceName := fmt.Sprintf("instance-%d", i+1)

		instanceIP, err := p.createInstance(ctx, instanceName, subnetURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create instance %s: %w", instanceName, err)
		}

		infra.InstanceNames = append(infra.InstanceNames, instanceName)
		infra.InstanceIPs = append(infra.InstanceIPs, instanceIP)

		log.Printf("Created instance %s with IP %s", instanceName, instanceIP)
	}

	return infra, nil
}

func (p *GCPProvisioner) createNetwork(ctx context.Context) (string, error) {
	req := &computepb.InsertNetworkRequest{
		Project: p.config.ProjectID,
		NetworkResource: &computepb.Network{
			Name:                  proto.String(p.config.NetworkName),
			AutoCreateSubnetworks: proto.Bool(false),
			RoutingConfig: &computepb.NetworkRoutingConfig{
				RoutingMode: proto.String("REGIONAL"),
			},
		},
	}

	op, err := p.networksClient.Insert(ctx, req)
	if err != nil {
		return "", err
	}

	if err := op.Wait(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("projects/%s/global/networks/%s", p.config.ProjectID, p.config.NetworkName), nil
}

func (p *GCPProvisioner) createSubnet(ctx context.Context, networkURL string) (string, error) {
	req := &computepb.InsertSubnetworkRequest{
		Project: p.config.ProjectID,
		Region:  p.config.Region,
		SubnetworkResource: &computepb.Subnetwork{
			Name:        proto.String(p.config.SubnetName),
			Network:     proto.String(networkURL),
			IpCidrRange: proto.String(p.config.SubnetCIDR),
			Region:      proto.String(p.config.Region),
		},
	}

	op, err := p.subnetworksClient.Insert(ctx, req)
	if err != nil {
		return "", err
	}

	if err := op.Wait(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", p.config.ProjectID, p.config.Region, p.config.SubnetName), nil
}

func (p *GCPProvisioner) createFirewallRule(ctx context.Context, networkURL string) (string, error) {
	req := &computepb.InsertFirewallRequest{
		Project: p.config.ProjectID,
		FirewallResource: &computepb.Firewall{
			Name:    proto.String(p.config.FirewallName),
			Network: proto.String(networkURL),
			Allowed: []*computepb.Allowed{
				{
					IPProtocol: proto.String("tcp"),
					Ports:      []string{"22", "80", "443"},
				},
			},
			SourceRanges: []string{"0.0.0.0/0"},
			Direction:    proto.String("INGRESS"),
		},
	}

	op, err := p.firewallsClient.Insert(ctx, req)
	if err != nil {
		return "", err
	}

	if err := op.Wait(ctx); err != nil {
		return "", err
	}

	return fmt.Sprintf("projects/%s/global/firewalls/%s", p.config.ProjectID, p.config.FirewallName), nil
}

func (p *GCPProvisioner) createInstance(ctx context.Context, name, subnetURL string) (string, error) {
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", p.config.Zone, p.config.MachineType)
	sourceImage := fmt.Sprintf("projects/ubuntu-os-cloud/global/images/family/%s", p.config.BootDiskImage)

	req := &computepb.InsertInstanceRequest{
		Project: p.config.ProjectID,
		Zone:    p.config.Zone,
		InstanceResource: &computepb.Instance{
			Name:        proto.String(name),
			MachineType: proto.String(machineType),
			Disks: []*computepb.AttachedDisk{
				{
					Boot:       proto.Bool(true),
					AutoDelete: proto.Bool(true),
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						SourceImage: proto.String(sourceImage),
						DiskSizeGb:  proto.Int64(p.config.BootDiskSizeGB),
						DiskType:    proto.String(fmt.Sprintf("zones/%s/diskTypes/pd-standard", p.config.Zone)),
					},
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Subnetwork: proto.String(subnetURL),
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name:        proto.String("External NAT"),
							Type:        proto.String("ONE_TO_ONE_NAT"),
							NetworkTier: proto.String("PREMIUM"),
						},
					},
				},
			},
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   proto.String("ssh-keys"),
						Value: proto.String(fmt.Sprintf("ubuntu:%s", p.config.SSHKey)),
					},
				},
			},
		},
	}

	op, err := p.instancesClient.Insert(ctx, req)
	if err != nil {
		return "", err
	}

	if err := op.Wait(ctx); err != nil {
		return "", err
	}

	instance, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  p.config.ProjectID,
		Zone:     p.config.Zone,
		Instance: name,
	})
	if err != nil {
		return "", err
	}

	if len(instance.NetworkInterfaces) > 0 && len(instance.NetworkInterfaces[0].AccessConfigs) > 0 {
		return *instance.NetworkInterfaces[0].AccessConfigs[0].NatIP, nil
	}

	return "", fmt.Errorf("no external IP found for instance %s", name)
}

func (p *GCPProvisioner) TeardownInfrastructure(ctx context.Context, infra *GCPInfrastructure) error {
	for _, instanceName := range infra.InstanceNames {
		if err := p.deleteInstance(ctx, instanceName); err != nil {
			log.Printf("Warning: failed to delete instance %s: %v", instanceName, err)
		}
	}

	if infra.FirewallURL != "" {
		if err := p.deleteFirewall(ctx); err != nil {
			log.Printf("Warning: failed to delete firewall: %v", err)
		}
	}

	if infra.SubnetURL != "" {
		if err := p.deleteSubnet(ctx); err != nil {
			log.Printf("Warning: failed to delete subnet: %v", err)
		}
	}

	if infra.NetworkURL != "" {
		if err := p.deleteNetwork(ctx); err != nil {
			return fmt.Errorf("failed to delete network: %w", err)
		}
	}

	return nil
}

func (p *GCPProvisioner) deleteInstance(ctx context.Context, name string) error {
	op, err := p.instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
		Project:  p.config.ProjectID,
		Zone:     p.config.Zone,
		Instance: name,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (p *GCPProvisioner) deleteFirewall(ctx context.Context) error {
	op, err := p.firewallsClient.Delete(ctx, &computepb.DeleteFirewallRequest{
		Project:  p.config.ProjectID,
		Firewall: p.config.FirewallName,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (p *GCPProvisioner) deleteSubnet(ctx context.Context) error {
	op, err := p.subnetworksClient.Delete(ctx, &computepb.DeleteSubnetworkRequest{
		Project:    p.config.ProjectID,
		Region:     p.config.Region,
		Subnetwork: p.config.SubnetName,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (p *GCPProvisioner) deleteNetwork(ctx context.Context) error {
	op, err := p.networksClient.Delete(ctx, &computepb.DeleteNetworkRequest{
		Project: p.config.ProjectID,
		Network: p.config.NetworkName,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (p *GCPProvisioner) ScaleInstances(ctx context.Context, infra *GCPInfrastructure, targetCount int) error {
	currentCount := len(infra.InstanceNames)

	if targetCount <= currentCount {
		return fmt.Errorf("target count %d must be greater than current count %d", targetCount, currentCount)
	}

	additionalCount := targetCount - currentCount

	for i := 0; i < additionalCount; i++ {
		instanceIndex := currentCount + i + 1
		instanceName := fmt.Sprintf("instance-%d", instanceIndex)

		instanceIP, err := p.createInstance(ctx, instanceName, infra.SubnetURL)
		if err != nil {
			return fmt.Errorf("failed to create instance %s: %w", instanceName, err)
		}

		infra.InstanceNames = append(infra.InstanceNames, instanceName)
		infra.InstanceIPs = append(infra.InstanceIPs, instanceIP)

		log.Printf("Created instance %s with IP %s (%d/%d)", instanceName, instanceIP, i+1, additionalCount)
	}

	return nil
}

func (p *GCPProvisioner) ListInstances(ctx context.Context) ([]*computepb.Instance, error) {
	req := &computepb.ListInstancesRequest{
		Project: p.config.ProjectID,
		Zone:    p.config.Zone,
	}

	it := p.instancesClient.List(ctx, req)
	var instances []*computepb.Instance

	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func (p *GCPProvisioner) GetInstanceDetails(ctx context.Context, instanceName string) (*computepb.Instance, error) {
	instance, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  p.config.ProjectID,
		Zone:     p.config.Zone,
		Instance: instanceName,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (p *GCPProvisioner) StopInstance(ctx context.Context, instanceName string) error {
	op, err := p.instancesClient.Stop(ctx, &computepb.StopInstanceRequest{
		Project:  p.config.ProjectID,
		Zone:     p.config.Zone,
		Instance: instanceName,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (p *GCPProvisioner) StartInstance(ctx context.Context, instanceName string) error {
	op, err := p.instancesClient.Start(ctx, &computepb.StartInstanceRequest{
		Project:  p.config.ProjectID,
		Zone:     p.config.Zone,
		Instance: instanceName,
	})
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}
