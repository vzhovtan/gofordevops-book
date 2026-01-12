package main

import "model"

func modelCRUD() {
	model, err := model.LoadModel("model.json")
	if err != nil {
		fmt.Printf("Error loading model: %v\n", err)
		return
	}

	fmt.Printf("Loaded modelstructure model version %s\n", model.Metadata.Version)
	fmt.Printf("Environment: %s\n", model.Metadata.Environment)
	fmt.Printf("Total devices: %d\n\n", len(model.Devices))

	device, err := model.GetDeviceByID(model, "rtr-core-01")
	if err != nil {
		fmt.Printf("Error getting device: %v\n", err)
		return
	}
	fmt.Printf("Found device: %s (%s %s)\n", device.Hostname, device.Vendor, device.Model)
	fmt.Printf("Management IP: %s\n", device.ManagementIP)
	fmt.Printf("Interfaces: %d\n\n", len(device.Interfaces))

	updates := map[string]interface{}{
		"description": "Updated uplink to core switch",
		"mtu":         9216,
	}

	err = model.UpdateDeviceInterface(model, "rtr-core-01", "GigabitEthernet0/0/0", updates)
	if err != nil {
		fmt.Printf("Error updating interface: %v\n", err)
		return
	}
	fmt.Println("Interface updated successfully")

	newRoute := model.StaticRoute{
		Destination:            "10.10.0.0/16",
		NextHop:                "192.168.10.2",
		AdministrativeDistance: 5,
	}

	err = model.AddStaticRoute(model, "rtr-core-01", newRoute)
	if err != nil {
		fmt.Printf("Error adding static route: %v\n", err)
		return
	}
	fmt.Println("Static route added successfully")

	err = model.UpdateDeviceManagementIP(model, "sw-access-01", "10.0.1.25")
	if err != nil {
		fmt.Printf("Error updating management IP: %v\n", err)
		return
	}
	fmt.Println("Management IP updated successfully")

	ciscoDevices := model.ListDevicesByVendor(model, "cisco")
	fmt.Printf("Cisco devices: %d\n", len(ciscoDevices))
	for _, d := range ciscoDevices {
		fmt.Printf("  - %s (%s)\n", d.Hostname, d.ID)
	}

	err = model.SaveModel("modelstructure_updated.json", model)
	if err != nil {
		fmt.Printf("Error saving model: %v\n", err)
		return
	}
	fmt.Println("\nModel saved successfully to modelstructure_updated.json")
}




func main() {
	modelCRUD()
}