package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// struct Device represents a network device with hostname and IPv4 address
type Device struct {
	Hostname    string `json:"hostname"`
	IPv4Address string `json:"ipv4_address"`
}

// InfraDevices represents the root structure of the JSON file
type InfraDevices struct {
	Devices []Device `json:"infrastructure_devices"`
}

// ReadDevicesFromFile reads the JSON file and returns a list of devices
func ReadDevicesFromFile(filename string) ([]Device, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse JSON
	var infraDevices InfraDevices
	err = json.Unmarshal(data, &infraDevices)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return infraDevices.Devices, nil
}

func Inventory() {
	// Read devices from JSON file
	devices, err := ReadDevicesFromFile("inventory.json")
	if err != nil {
		log.Fatal(err)
	}

	// Print the devices
	fmt.Printf("Found %d infrastructure devices:\n\n", len(devices))
	for _, device := range devices {
		fmt.Printf("%-20s %s\n", device.Hostname, device.IPv4Address)
	}
}
