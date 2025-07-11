package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type USBDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"device_path"`
}

type USBDeviceData struct {
	AvailableUSBDevices []USBDevice `json:"available_midi_ports"`
	SelectedUSBDevice   string      `json:"selected_midi_port"`
}

type MIDIDevice struct {
	Name     string `json:"name"`
	PortPath string `json:"port_path"`
}

type MIDIDeviceData struct {
	AvailableMIDIDevices []MIDIDevice `json:"available_midi_ports"`
	SelectedMIDIDevice   MIDIDevice   `json:"selected_midi_port"`
}

var (
	rootPath           string = getRootPath()
	confPath           string = filepath.Join(rootPath, "backend", "modularMidi.conf") // Edited rootPath to lead to .conf File
	backendApiLocation string = generateBackendApiLocation()
)

func generateBackendApiLocation() string {
	httpconfRAW := loadHTTPconf()
	return strings.Join([]string{parseProtocol(httpconfRAW), "://", parseHost(httpconfRAW), ":", parsePort(httpconfRAW)}, "")
}

func loadHTTPconf() string {
	fmt.Println("Loading configuration from: \n", confPath) // Debug output
	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	getKey := func(section, key string) string {
		s := cfg.Section(section)
		if !s.HasKey(key) {
			log.Fatalf("Missing key [%s] %s", section, key)
		}
		return s.Key(key).String()
	}

	returnStr := strings.Join([]string{
		"listen_port:",
		getKey("http", "listen_port"),
		",backend_api_port:",
		getKey("http", "backend_api_port"),
		",backend_api_host:",
		getKey("http", "backend_api_host"),
		",backend_api_protocol:",
		getKey("http", "backend_api_protocol"),
	}, "")
	return returnStr
}

func parseHost(unparsed string) string {
	var host string
	parts := strings.Split(unparsed, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "backend_api_host:") {
			host = strings.TrimPrefix(part, "backend_api_host:")
			break
		}
	}
	return host
}

func parsePort(unparsed string) string {
	var port string
	parts := strings.Split(unparsed, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "listen_port:") {
			port = strings.TrimPrefix(part, "listen_port:")
			break
		}
	}
	return port
}

func parseProtocol(unparsed string) string {
	var protocol string
	parts := strings.Split(unparsed, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "backend_api_protocol:") {
			protocol = strings.TrimPrefix(part, "backend_api_protocol:")
			break
		}
	}
	fmt.Println(protocol)
	return protocol
}

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list-USB":
		listUSBDevices()
	case "select-USB":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a USB device index to select")
			fmt.Println("Usage: usb-manager select <index>")
			os.Exit(1)
		}
		selectUSBDevice(os.Args[2])
	case "select-MIDI":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a MIDI device index to select")
			fmt.Println("Usage: usb-manager select-MIDI <index>")
			os.Exit(1)
		}
		selectMIDIDevice(os.Args[2])
	case "test-midi":
		testMidiOutput()
		fmt.Println("MIDI output test triggered.")
	case "list-midi":
		listMIDI()
		fmt.Println("MIDI ports listed.")
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

func getRootPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	parentDir := filepath.Dir(dir)
	return parentDir
}

// getUSBFileContent retrieves the USB device data from the API and reads the JSON file
func getUSBFileContent() (*USBDeviceData, string, error) {

	resp, err := http.Get(strings.Join([]string{backendApiLocation, "/usbPortListFile"}, ""))
	fmt.Println("Response: ", resp)
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

//getMIDIFileContent

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

func testMidiOutput() {
	fmt.Print("Backend Loc: ")
	fmt.Println(backendApiLocation)
	resp, err := http.Get(strings.Join([]string{backendApiLocation, "/testMidiOutput"}, ""))
	if err != nil {
		fmt.Printf("Failed to call API: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned status code: %d\n", resp.StatusCode)
	}
}

func getMIDIFileContent() (*MIDIDeviceData, string, error) {
	resp, err := http.Get(strings.Join([]string{backendApiLocation, "/listMidiPorts"}, ""))
	if err != nil {
		return nil, "", fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Response: ", resp)
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
	var midiData MIDIDeviceData
	err = json.Unmarshal(fileContent, &midiData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &midiData, filePath, nil
}

func listMIDI() {
	fmt.Print("Backend Loc: ")
	fmt.Println(backendApiLocation)
	resp, err := http.Get(strings.Join([]string{backendApiLocation, "/listMidiPorts"}, ""))
	if err != nil {
		fmt.Printf("Failed to call API: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned status code: %d\n", resp.StatusCode)
	}
	midiData, _, err := getMIDIFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available MIDI Devices:")
	fmt.Println("======================")

	if len(midiData.AvailableMIDIDevices) == 0 {
		fmt.Println("No MIDI devices found.")
		return
	}

	for i, device := range midiData.AvailableMIDIDevices {
		fmt.Printf("[%d] %s\n", i+1, device.Name)
		fmt.Printf("    Device Path: %s\n", device.PortPath)
		fmt.Println()
	}

	if midiData.SelectedMIDIDevice.PortPath != "" {
		fmt.Printf("Currently selected MIDI device path: %s\n", midiData.SelectedMIDIDevice.PortPath)
	} else {
		fmt.Println("No MIDI device currently selected.")
	}
}

func selectUSBDevice(indexStr string) {
	usbData, filePath, err := getUSBFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("Backend Loc: ")

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

func selectMIDIDevice(indexStr string) {
	midiData, filePath, err := getMIDIFileContent()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("Backend Loc: ")

	// Parse the index
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		fmt.Printf("Error: Invalid index '%s'. Please provide a valid number.\n", indexStr)
		os.Exit(1)
	}

	// Validate the index
	if index < 1 || index > len(midiData.AvailableMIDIDevices) {
		fmt.Printf("Error: Index %d is out of range. Available devices: 1-%d\n", index, len(midiData.AvailableMIDIDevices))
		os.Exit(1)
	}

	// Get the selected device (convert to 0-based index)
	selectedDevice := midiData.AvailableMIDIDevices[index-1]

	// Update the selected MIDI device in the data structure (store the whole struct)
	midiData.SelectedMIDIDevice = selectedDevice

	// Marshal the updated data back to JSON with proper formatting
	updatedJSON, err := json.MarshalIndent(midiData, "", "  ")
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

	fmt.Printf("Successfully selected MIDI device:\n")
	fmt.Printf("  Name: %s\n", selectedDevice.Name)
	fmt.Printf("  Device Path: %s\n", selectedDevice.PortPath)
	fmt.Printf("  Saved to: %s\n", filePath)
}
