package usbUtility

import (
	"encoding/json"
	"fmt"
	"log"
	getvalues "modularMidiGoApp/backend/getValues"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type USBPort struct {
	Bus    string `json:"bus"`
	Device string `json:"device"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

var (
	rootPath = getvalues.FindRootPath() // Gets root path from getValues package
	dirPath  = filepath.Join(rootPath, "usbUtility")
	filePath = filepath.Join(dirPath, "usb_ports.json")
)

// UsbPortLists retrieves the list of USB ports and writes them to a JSON file.
// It uses the `lsusb` command to gather information about connected USB devices.

func UsbPortLists() string {
	cmd := exec.Command("lsusb")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute lsusb: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var ports []USBPort

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}
		bus := parts[1]
		device := strings.TrimSuffix(parts[3], ":")
		id := parts[5]
		name := strings.Join(parts[6:], " ")
		ports = append(ports, USBPort{
			Bus:    bus,
			Device: device,
			ID:     id,
			Name:   name,
		})
	}

	jsonData, err := json.MarshalIndent(ports, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println("available USB Ports:")
	fmt.Println(string(jsonData))
	writeToFile(rootPath, jsonData)
	return filePath
}

func writeToFile(rootPath string, data []byte) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	var fileData map[string]interface{}
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err == nil {
			json.Unmarshal(content, &fileData)
		}
	}
	if fileData == nil {
		fileData = make(map[string]interface{})
	}
	var ports []interface{}
	json.Unmarshal(data, &ports)
	fileData["available_usb_ports"] = ports
	finalData, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal final JSON: %w", err)
	}
	data = finalData

	return os.WriteFile(filePath, data, 0644)
}
