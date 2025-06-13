package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type USBDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"device_path"`
}

type USBDeviceData struct {
	AvailableUSBDevices []USBDevice `json:"available_usb_devices"`
	SelectedUSBDevice   string      `json:"selected_usb_device"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listUSBDevices()
	case "select":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a USB device index to select")
			fmt.Println("Usage: usb-manager select <index>")
			os.Exit(1)
		}
		selectUSBDevice(os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() { // Print usage instructions for the CLI tool
	fmt.Println("USB Device Manager CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  usb-manager list           - List all available USB devices")
	fmt.Println("  usb-manager select <index> - Select a USB device by index")
	fmt.Println("  usb-manager help           - Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  usb-manager list")
	fmt.Println("  usb-manager select 3")
}

// getUSBFileContent retrieves the USB device data from the API and reads the JSON file
func getUSBFileContent() (*USBDeviceData, string, error) {
	// Get the file path from the API
	resp, err := http.Get("http://localhost:18181/usbPortListFile")
	if err != nil {
		return nil, "", fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read API response: %v", err)
	}

	filePath := strings.TrimSpace(string(body))
	if filePath == "" {
		return nil, "", fmt.Errorf("API returned empty file path")
	}

	// Read the JSON file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	//fmt.Print(string(fileContent))
	// Parse the JSON content
	var usbData USBDeviceData
	err = json.Unmarshal(fileContent, &usbData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &usbData, filePath, nil
}

func listUSBDevices() {
	usbData, _, err := getUSBFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available USB Devices:")
	fmt.Println("=====================")

	if len(usbData.AvailableUSBDevices) == 0 {
		fmt.Println("No USB devices found.")
		return
	}

	for i, device := range usbData.AvailableUSBDevices {
		fmt.Printf("[%d] %s\n", i+1, device.Name)
		fmt.Printf("    Device Path: %s\n", device.DevicePath)
		fmt.Println()
	}

	if usbData.SelectedUSBDevice != "" {
		fmt.Printf("Currently selected USB device path: %s\n", usbData.SelectedUSBDevice)
	} else {
		fmt.Println("No USB device currently selected.")
	}
}

func selectUSBDevice(indexStr string) {
	usbData, filePath, err := getUSBFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Parse the index
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		fmt.Printf("Error: Invalid index '%s'. Please provide a valid number.\n", indexStr)
		os.Exit(1)
	}

	// Validate the index
	if index < 1 || index > len(usbData.AvailableUSBDevices) {
		fmt.Printf("Error: Index %d is out of range. Available devices: 1-%d\n", index, len(usbData.AvailableUSBDevices))
		os.Exit(1)
	}

	// Get the selected device (convert to 0-based index)
	selectedDevice := usbData.AvailableUSBDevices[index-1]

	// Update the selected USB device in the data structure (store device path)
	usbData.SelectedUSBDevice = selectedDevice.DevicePath

	// Marshal the updated data back to JSON with proper formatting
	updatedJSON, err := json.MarshalIndent(usbData, "", "  ")
	if err != nil {
		fmt.Printf("Error: Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	// Write the updated data back to the file
	err = os.WriteFile(filePath, updatedJSON, 0644)
	if err != nil {
		fmt.Printf("Error: Failed to write to file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully selected USB device:\n")
	fmt.Printf("  Name: %s\n", selectedDevice.Name)
	fmt.Printf("  Device Path: %s\n", selectedDevice.DevicePath)
	fmt.Printf("  Saved to: %s\n", filePath)
}
