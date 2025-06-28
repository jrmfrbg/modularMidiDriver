package usbUtility

import (
	"encoding/json"
	"fmt"
	"log"
	getvalues "modularMidiGoApp/backend/getValues"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type USBDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"device_path"`
}

var (
	rootPath = getvalues.FindRootPath() // Gets root path from getValues package
	dirPath  = filepath.Join(rootPath, "usbUtility")
	FilePath = filepath.Join(dirPath, "usb_ports.json")
)

// UsbPortLists retrieves the list of USB devices with their names and device paths.
func UsbPortLists() string {
	fmt.Printf("Running on OS: %s\n", runtime.GOOS) // Debug output

	devices := findUSBDevices()

	// Ensure we always have a valid slice (not nil)
	if devices == nil {
		devices = make([]USBDevice, 0)
	}

	jsonData, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println("Available USB Devices:")
	fmt.Println(string(jsonData))
	writeToFile(jsonData)
	return FilePath
}

// findUSBDevices finds USB devices and their corresponding device paths
func findUSBDevices() []USBDevice {
	devices := make([]USBDevice, 0) // Initialize empty slice

	// Find all serial-like device paths
	devicePaths := findSerialDevices()

	fmt.Printf("Found %d device paths: %v\n", len(devicePaths), devicePaths) // Debug output

	for _, path := range devicePaths {
		name := getDeviceName(path)
		fmt.Printf("Device path: %s, Name: %s\n", path, name) // Debug output

		// Always add the device, even if name is empty (will use path as fallback)
		if name == "" {
			name = filepath.Base(path)
		}

		devices = append(devices, USBDevice{
			Name:       name,
			DevicePath: path,
		})
	}

	fmt.Printf("Total devices found: %d\n", len(devices)) // Debug output
	return devices
}

// findSerialDevices finds all serial device paths based on OS
func findSerialDevices() []string {
	switch runtime.GOOS {
	case "windows":
		return findWindowsSerialDevices()
	case "darwin":
		return findMacOSSerialDevices()
	case "linux":
		return findLinuxSerialDevices()
	default:
		return findLinuxSerialDevices() // Default to Linux patterns
	}
}

// findWindowsSerialDevices finds Windows COM ports
func findWindowsSerialDevices() []string {
	var paths []string

	// Use wmic to get COM port information
	cmd := exec.Command("wmic", "path", "Win32_SerialPort", "get", "DeviceID", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try common COM port range
		for i := 1; i <= 20; i++ {
			comPort := fmt.Sprintf("COM%d", i)
			if isWindowsPortAvailable(comPort) {
				paths = append(paths, comPort)
			}
		}
		return paths
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && strings.HasPrefix(parts[1], "COM") {
			comPort := strings.TrimSpace(parts[1])
			paths = append(paths, comPort)
		}
	}

	return paths
}

// findMacOSSerialDevices finds macOS USB serial devices
func findMacOSSerialDevices() []string {
	var paths []string

	patterns := []string{
		"/dev/cu.usb*",
		"/dev/cu.usbmodem*",
		"/dev/cu.usbserial*",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		paths = append(paths, matches...)
	}

	return paths
}

// findLinuxSerialDevices finds Linux USB serial devices
func findLinuxSerialDevices() []string {
	var paths []string

	patterns := []string{
		"/dev/ttyUSB*",
		"/dev/ttyACM*",
		"/dev/ttyS*", // Include regular serial ports too
	}

	fmt.Printf("Checking Linux patterns: %v\n", patterns) // Debug output

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Error with pattern %s: %v\n", pattern, err)
			continue
		}
		fmt.Printf("Pattern %s found: %v\n", pattern, matches) // Debug output
		paths = append(paths, matches...)
	}

	return paths
}

// isWindowsPortAvailable checks if a Windows COM port exists
func isWindowsPortAvailable(port string) bool {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Get-WmiObject -Class Win32_SerialPort | Where-Object {$_.DeviceID -eq '%s'}", port))
	output, err := cmd.Output()
	return err == nil && len(strings.TrimSpace(string(output))) > 0
}

// getDeviceName attempts to get a human-readable name for the device based on OS
func getDeviceName(devicePath string) string {
	switch runtime.GOOS {
	case "windows":
		return getWindowsDeviceName(devicePath)
	case "darwin":
		return getMacOSDeviceName(devicePath)
	case "linux":
		return getLinuxDeviceName(devicePath)
	default:
		return getLinuxDeviceName(devicePath)
	}
}

// getWindowsDeviceName gets device name on Windows
func getWindowsDeviceName(comPort string) string {
	// Method 1: Use wmic to get device description
	cmd := exec.Command("wmic", "path", "Win32_SerialPort", "where", fmt.Sprintf("DeviceID='%s'", comPort), "get", "Description", "/format:csv")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			parts := strings.Split(line, ",")
			if len(parts) >= 2 && parts[1] != "Description" && strings.TrimSpace(parts[1]) != "" {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// Method 2: Use PowerShell to get friendly name
	cmd = exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-WmiObject -Class Win32_SerialPort | Where-Object {$_.DeviceID -eq '%s'} | Select-Object -ExpandProperty Name", comPort))
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return strings.TrimSpace(string(output))
	}

	// Method 3: Try to get PnP device info
	cmd = exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-WmiObject -Class Win32_PnPEntity | Where-Object {$_.Name -like '*%s*'} | Select-Object -ExpandProperty Name", comPort))
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return strings.TrimSpace(string(output))
	}

	// Fallback: return COM port name
	return comPort
}

// getMacOSDeviceName gets device name on macOS
func getMacOSDeviceName(devicePath string) string {
	// Try ioreg method
	if name := getNameFromIOReg(devicePath); name != "" {
		return name
	}

	// Fallback: use device path basename
	return filepath.Base(devicePath)
}

// getLinuxDeviceName gets device name on Linux
func getLinuxDeviceName(devicePath string) string {
	// Method 1: Try udevadm
	if name := getNameFromUdev(devicePath); name != "" {
		return name
	}

	// Method 2: Try lsusb mapping
	if name := getNameFromLsusb(devicePath); name != "" {
		return name
	}

	// Fallback: use device path basename
	return filepath.Base(devicePath)
}

// getNameFromUdev gets device name using udevadm (Linux)
func getNameFromUdev(devicePath string) string {
	cmd := exec.Command("udevadm", "info", "--name="+devicePath, "--query=property")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID_MODEL=") {
			return strings.TrimPrefix(line, "ID_MODEL=")
		}
		if strings.HasPrefix(line, "ID_MODEL_FROM_DATABASE=") {
			return strings.TrimPrefix(line, "ID_MODEL_FROM_DATABASE=")
		}
	}

	return ""
}

// getNameFromIOReg gets device name using ioreg (macOS)
func getNameFromIOReg(devicePath string) string {
	// Extract the device identifier from path (e.g., cu.usbmodem14101 -> usbmodem14101)
	baseName := filepath.Base(devicePath)
	if strings.HasPrefix(baseName, "cu.") {
		baseName = strings.TrimPrefix(baseName, "cu.")
	}

	cmd := exec.Command("ioreg", "-p", "IOUSB", "-l")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	outputStr := string(output)

	// Look for USB device entries that might match
	re := regexp.MustCompile(`"USB Product Name" = "([^"]+)"`)
	matches := re.FindAllStringSubmatch(outputStr, -1)

	if len(matches) > 0 {
		// Return the first match (you might want to improve this logic)
		return matches[0][1]
	}

	return ""
}

// getNameFromLsusb attempts to map device path to lsusb output
func getNameFromLsusb(devicePath string) string {
	// Get USB device info via udevadm first to find bus/device numbers
	cmd := exec.Command("udevadm", "info", "--name="+devicePath, "--attribute-walk")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	outputStr := string(output)

	// Extract bus and device numbers
	busRe := regexp.MustCompile(`ATTRS{busnum}=="(\d+)"`)
	deviceRe := regexp.MustCompile(`ATTRS{devnum}=="(\d+)"`)

	busMatches := busRe.FindStringSubmatch(outputStr)
	deviceMatches := deviceRe.FindStringSubmatch(outputStr)

	if len(busMatches) < 2 || len(deviceMatches) < 2 {
		return ""
	}

	busNum := fmt.Sprintf("%03s", busMatches[1])
	deviceNum := fmt.Sprintf("%03s", deviceMatches[1])

	// Now get lsusb output and find matching device
	cmd = exec.Command("lsusb")
	output, err = cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Bus "+busNum) && strings.Contains(line, "Device "+deviceNum) {
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				return strings.Join(parts[6:], " ")
			}
		}
	}

	return ""
}

func writeToFile(data []byte) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var fileData map[string]interface{}
	if _, err := os.Stat(FilePath); err == nil {
		content, err := os.ReadFile(FilePath)
		if err == nil {
			json.Unmarshal(content, &fileData)
		}
	}

	if fileData == nil {
		fileData = make(map[string]interface{})
	}

	var devices []interface{}
	json.Unmarshal(data, &devices)
	fileData["available_usb_devices"] = devices

	finalData, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return os.WriteFile(FilePath, finalData, 0644)
}
