package usbUtility

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.bug.st/serial"
)

var USBListenerStop = make(chan struct{})

// USBPortsList represents the overall structure of the JSON file.
type USBPortsList struct {
	AvailableUSBDevices []USBDevice `json:"available_usb_devices"`
	SelectedUSBDevice   string      `json:"selected_usb_device"`
}

func USBListener() {
	// List available serial ports
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatalf("Error getting port list: %v", err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}

	fmt.Println("Available serial ports:")
	for _, port := range ports {
		fmt.Printf("- %s\n", port)
	}

	// Try to automatically find the ESP32 port.
	// This often involves looking for specific names, but it can be
	// tricky to make universally reliable.
	// For example, on Linux, it might be something like "/dev/ttyACM0" or "/dev/ttyUSB0".
	// On Windows, "COMx".
	// The ESP32's native USB CDC often appears as "ACM" or a similar name.
	var esp32Port string
	for _, p := range ports {
		// You might need to adjust this heuristic based on your OS and ESP32 model.
		// For ESP32 native USB, it's often /dev/ttyACM0 on Linux.
		// On Windows, it could be "COMx" with a specific device description.
		selectedUSBDevice, err := getSelectedUSBDevice(FilePath)
		fmt.Printf("Selected USB Device: %s\n", selectedUSBDevice)
		if err != nil {
			log.Println("Error getting selected USB device: %v", err)
		}

		if strings.Contains(p, selectedUSBDevice) { // Basic heuristic
			fmt.Printf("Attempting to connect to: %s (heuristic match)\n", p)
			esp32Port = p
			break
		}
	}

	if esp32Port == "" {
		log.Fatal("Could not automatically find a likely ESP32 serial port. Please specify manually.")
		// If you can't find it automatically, uncomment the line below and set it manually:
		// esp32Port = "/dev/ttyACM0" // Or "COM3" on Windows
	}

	// Open the serial port
	mode := &serial.Mode{
		BaudRate: 115200, // This baud rate *does* matter for the Go client
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(esp32Port, mode)
	if err != nil {
		log.Fatalf("Error opening serial port %s: %v", esp32Port, err)
	}
	defer port.Close()

	fmt.Printf("Successfully opened serial port: %s\n", esp32Port)

	// --- GoRoutine for reading from serial port ---
	go func() {
		scanner := bufio.NewScanner(port)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("[ESP32] %s\n", line)
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from serial port: %v\n", err)
		}
	}()

	// --- Main loop for sending data to serial port ---
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter text to send to ESP32 (type 'exit' to quit):")
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input) // Remove newline and carriage return

		if input == "exit" {
			fmt.Println("Exiting client.")
			return
		}

		// Send the input to ESP32, adding a newline for the ESP32 to recognize it
		_, err := port.Write([]byte(input + "\n"))
		if err != nil {
			log.Printf("Error writing to serial port: %v\n", err)
			continue
		}
		time.Sleep(10 * time.Millisecond) // Give a tiny delay
	}
}

func getSelectedUSBDevice(usbPortsListFile string) (string, error) {
	// Read the content of the JSON file.
	fileContent, err := os.ReadFile(usbPortsListFile)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", usbPortsListFile, err)
	}

	// Create an instance of the USBPortsList struct to hold the unmarshaled data.
	var config USBPortsList

	// Unmarshal the JSON content into the config struct.
	if err := json.Unmarshal(fileContent, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON from '%s': %w", usbPortsListFile, err)
	}

	// Return the selected USB device.
	return config.SelectedUSBDevice, nil
}
