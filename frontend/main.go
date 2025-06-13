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

type USBPort struct {
	Bus    string `json:"bus"`
	Device string `json:"device"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

type USBPortData struct {
	AvailableUSBPorts []USBPort `json:"available_usb_ports"`
	SelectedUSBPort   string    `json:"selected_usb_port"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listUSBPorts()
	case "select":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a USB port index to select")
			fmt.Println("Usage: usb-manager select <index>")
			os.Exit(1)
		}
		selectUSBPort(os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("USB Port Manager CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  usb-manager list           - List all available USB ports")
	fmt.Println("  usb-manager select <index> - Select a USB port by index")
	fmt.Println("  usb-manager help           - Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  usb-manager list")
	fmt.Println("  usb-manager select 3")
}

func getUSBFileContent() (*USBPortData, string, error) {
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

	// Parse the JSON content
	var usbData USBPortData
	err = json.Unmarshal(fileContent, &usbData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &usbData, filePath, nil
}

func listUSBPorts() {
	usbData, _, err := getUSBFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available USB Ports:")
	fmt.Println("===================")

	if len(usbData.AvailableUSBPorts) == 0 {
		fmt.Println("No USB ports found.")
		return
	}

	for i, port := range usbData.AvailableUSBPorts {
		fmt.Printf("[%d] Bus: %s, Device: %s, ID: %s\n", i+1, port.Bus, port.Device, port.ID)
		fmt.Printf("    Name: %s\n", port.Name)
		fmt.Println()
	}

	if usbData.SelectedUSBPort != "" {
		fmt.Printf("Currently selected USB port: %s\n", usbData.SelectedUSBPort)
	} else {
		fmt.Println("No USB port currently selected.")
	}
}

func selectUSBPort(indexStr string) {
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
	if index < 1 || index > len(usbData.AvailableUSBPorts) {
		fmt.Printf("Error: Index %d is out of range. Available ports: 1-%d\n", index, len(usbData.AvailableUSBPorts))
		os.Exit(1)
	}

	// Get the selected port (convert to 0-based index)
	selectedPort := usbData.AvailableUSBPorts[index-1]

	// Update the selected USB port in the data structure
	usbData.SelectedUSBPort = selectedPort.ID

	// Marshal the updated data back to JSON with proper formatting
	updatedJSON, err := json.MarshalIndent(usbData, "", " ")
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

	fmt.Printf("Successfully selected USB port:\n")
	fmt.Printf("  Bus: %s, Device: %s, ID: %s\n", selectedPort.Bus, selectedPort.Device, selectedPort.ID)
	fmt.Printf("  Name: %s\n", selectedPort.Name)
	fmt.Printf("  Saved to: %s\n", filePath)
}
