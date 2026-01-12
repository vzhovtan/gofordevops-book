package main

import (
	"push"
	"model"
	"render"
)

func fullPush() {
	model, err := model.LoadModel("model.json")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	device := &model.Devices[0]
	if device.Vendor != "cisco" {
		log.Fatalf("This example requires a Cisco device")
	}

	config, err := render.GenerateConfiguration(model, device.ID)
	if err != nil {
		log.Fatalf("Failed to generate configuration: %v", err)
	}

	strategy := push.NewFullReplaceStrategy("admin", "password", 30*time.Second)
	deployer := push.NewConfigDeployer(strategy)

	result := deployer.DeployToDevice(device, config)

	if result.Success {
		fmt.Printf("Deployment successful!\n")
		fmt.Printf("Device: %s\n", result.DeviceID)
		fmt.Printf("Duration: %v\n", result.Duration)
	} else {
		fmt.Printf("Deployment failed: %v\n", result.Error)
	}
}

func partialPush() {
	model, err := model.LoadModel("infrastructure.json")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	var juniperDevice *model.Device
	for i := range model.Devices {
		if model.Devices[i].Vendor == "juniper" {
			juniperDevice = &model.Devices[i]
			break
		}
	}

	if juniperDevice == nil {
		log.Fatalf("No Juniper device found in model")
	}

	strategy := push.NewPerElementStrategy("admin", "password", 30*time.Second)
	deployer := push.NewPerElementDeployer(strategy)

	updates := map[string]interface{}{
		"description": "Updated via per-element strategy",
		"mtu":         9000,
	}

	err = deployer.UpdateInterface(juniperDevice, "ge-0/0/0", updates)
	if err != nil {
		log.Fatalf("Failed to update interface: %v", err)
	}

	fmt.Println("Interface updated successfully")
}

func main() {
	fullPush()
	partialPush()
}