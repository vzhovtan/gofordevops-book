package crawl

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

type ModuleArgs struct {
    NetworkRanges     []string `json:"network_ranges"`
    ConcurrentDevices int      `json:"concurrent_devices"`
    DatabaseURL       string   `json:"database_url"`
    SSHKeyPath        string   `json:"ssh_key_path"`
    SSHUsername       string   `json:"ssh_username"`
    LogLevel          string   `json:"log_level"`
}

type ModuleResponse struct {
    Changed          bool              `json:"changed"`
    Failed           bool              `json:"failed,omitempty"`
    Msg              string            `json:"msg,omitempty"`
    DevicesDiscovered int              `json:"devices_discovered,omitempty"`
    DevicesByVendor  map[string]int    `json:"devices_by_vendor,omitempty"`
    CrawlDuration    float64           `json:"crawl_duration_seconds,omitempty"`
}

type ModuleInput struct {
    ANSIBLE_MODULE_ARGS ModuleArgs `json:"ANSIBLE_MODULE_ARGS"`
}

func parseArgs(filename string) (*ModuleArgs, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var input ModuleInput
    if err := json.Unmarshal(data, &input); err != nil {
        return nil, err
    }

    return &input.ANSIBLE_MODULE_ARGS, nil
}

func emitResponse(response ModuleResponse) {
    output, err := json.Marshal(response)
    if err != nil {
        failModule(fmt.Sprintf("Failed to marshal response: %v", err))
        return
    }
    fmt.Println(string(output))
}

func failModule(msg string) {
    response := ModuleResponse{
        Failed: true,
        Msg:    msg,
    }
    output, _ := json.Marshal(response)
    fmt.Println(string(output))
    os.Exit(1)
}

func executeCrawl(args *ModuleArgs) ModuleResponse {
    startTime := time.Now()
    
    // Configure crawler with provided parameters
    config := &CrawlerConfig{
        NetworkRanges:     args.NetworkRanges,
        ConcurrentDevices: args.ConcurrentDevices,
        DatabaseURL:       args.DatabaseURL,
        SSHKeyPath:        args.SSHKeyPath,
        SSHUsername:       args.SSHUsername,
        LogLevel:          args.LogLevel,
    }
    
    crawler, err := NewCrawler(config)
    if err != nil {
        return ModuleResponse{
            Failed: true,
            Msg:    fmt.Sprintf("Failed to initialize crawler: %v", err),
        }
    }
    
    // Execute discovery
    results, err := crawler.Discover()
    if err != nil {
        return ModuleResponse{
            Failed: true,
            Msg:    fmt.Sprintf("Crawl failed: %v", err),
        }
    }
    
    // Store results in database
    if err := crawler.StoreInventory(results); err != nil {
        return ModuleResponse{
            Failed: true,
            Msg:    fmt.Sprintf("Failed to store inventory: %v", err),
        }
    }
    
    duration := time.Since(startTime).Seconds()
    
    // Aggregate results by vendor
    devicesByVendor := make(map[string]int)
    for _, device := range results.Devices {
        devicesByVendor[device.Vendor]++
    }
    
    return ModuleResponse{
        Changed:           len(results.Devices) > 0,
        Msg:               fmt.Sprintf("Discovered %d devices across %d vendors", len(results.Devices), len(devicesByVendor)),
        DevicesDiscovered: len(results.Devices),
        DevicesByVendor:   devicesByVendor,
        CrawlDuration:     duration,
    }
}

func main() {
    if len(os.Args) != 2 {
        failModule("No argument file provided")
        return
    }

    argsFile := os.Args[1]
    args, err := parseArgs(argsFile)
    if err != nil {
        failModule(fmt.Sprintf("Failed to parse arguments: %v", err))
        return
    }

    response := executeCrawl(args)
    emitResponse(response)
}