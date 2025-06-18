package midiCCOutputer

import (
	"encoding/json"
	"fmt"
	getvalues "modularMidiGoApp/backend/getValues"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type USBDevice struct {
	Name     string `json:"name"`
	PortPath string `json:"port_path"`
}

var (
	rootPath = getvalues.FindRootPath()
	dirPath  = filepath.Join(rootPath, "midiUtility")
	filePath = filepath.Join(dirPath, "midi_ports.json")
)

func ListMIDIPorts() string {
	writeToFile(readMIDIPorts())
	return filePath
}

func readMIDIPorts() string {
	// Get available MIDI output ports
	outs := midi.GetOutPorts()

	if len(outs) == 0 {
		fmt.Print("No MIDI output ports available")
		return ""
	}

	fmt.Printf("Found %d MIDI output ports\n", len(outs)) // Debug output
	for i, out := range outs {
		fmt.Printf("Port %d: %s\n", i+1, out.String()) // Debug output
	}

	var portList string
	for i, out := range outs {
		portList += fmt.Sprintf("%d: %s\n", i+1, out.String())
	}

	return portList
}

func parseMIDIPorts(dataStr string) []USBDevice {
	var devices []USBDevice
	lines := strings.Split(strings.TrimSpace(dataStr), "\n")
	for _, line := range lines {
		// Example: "Port 1: Midi Through:Midi Through Port-0 14:0"
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		portInfo := parts[1]
		lastSpace := strings.LastIndex(portInfo, " ")
		if lastSpace == -1 {
			continue
		}
		name := portInfo[:lastSpace]
		portPath := portInfo[lastSpace+1:]
		devices = append(devices, USBDevice{
			Name:     name,
			PortPath: portPath,
		})
	}
	return devices
}

func writeToFile(dataStr string) error {
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

	devices := parseMIDIPorts(dataStr)
	fileData["available_midi_ports"] = devices

	finalData, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return os.WriteFile(filePath, finalData, 0644)
}
