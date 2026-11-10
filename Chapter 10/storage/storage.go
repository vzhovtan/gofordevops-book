package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// NetworkDevice represents a network device with its properties
type NetworkDevice struct {
	Hostname    string    `json:"hostname"`
	IPAddress   string    `json:"ip_address"`
	DeviceType  string    `json:"device_type"`
	LastChecked time.Time `json:"last_checked"`
	IsActive    bool      `json:"is_active"`
}

// InventoryData represents the complete inventory structure
type InventoryData struct {
	CollectionTime time.Time       `json:"collection_time"`
	TotalDevices   int             `json:"total_devices"`
	Devices        []NetworkDevice `json:"devices"`
}

// SaveToJSON saves data to a JSON file
func SaveToJSON(filename string, data interface{}) error {
	// Marshal data to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	// Write to file with appropriate permissions
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("Data successfully saved to: %s\n", filename)
	return nil
}

// AppendDeviceToFile appends a single device to the existing JSON file
func AppendDeviceToFile(filename string, device NetworkDevice) error {
	var inventory InventoryData

	data, err := os.ReadFile(filename)
	if err != nil {
		// File doesn't exist, create new inventory
		inventory = InventoryData{
			CollectionTime: time.Now(),
			Devices:        []NetworkDevice{},
		}
	} else {
		// Parse existing data
		err = json.Unmarshal(data, &inventory)
		if err != nil {
			return fmt.Errorf("error parsing existing JSON: %v", err)
		}
	}

	// Append the new device
	inventory.Devices = append(inventory.Devices, device)
	inventory.TotalDevices = len(inventory.Devices)
	inventory.CollectionTime = time.Now()

	// Save updated inventory
	return SaveToJSON(filename, inventory)
}

// SaveDeviceMap saves a map of devices to JSON
func SaveDeviceMap(filename string, devices map[string]NetworkDevice) error {
	jsonData, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Device map saved to: %s\n", filename)
	return nil
}
