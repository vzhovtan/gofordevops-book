package main

import "render"

func renderRunning() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <json-file> [device-id]")
		os.Exit(1)
	}

	filename := os.Args[1]

	model, err := render.LoadModel(filename)
	if err != nil {
		fmt.Printf("Error loading model: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded model with %d devices\n\n", len(model.Devices))

	if len(os.Args) >= 3 {
		deviceID := os.Args[2]
		config, err := render.GenerateConfiguration(model, deviceID)
		if err != nil {
			fmt.Printf("Error generating configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(config)
	} else {
		for _, device := range model.Devices {
			fmt.Printf("Generating configuration for %s (%s %s)\n", device.Hostname, device.Vendor, device.Model)
			fmt.Println(strings.Repeat("=", 80))

			config, err := render.GenerateConfiguration(model, device.ID)
			if err != nil {
				fmt.Printf("Error: %v\n\n", err)
				continue
			}

			fmt.Println(config)
			fmt.Println()
		}
	}
}

func main(){
	renderRunning()
}