package usbUtility

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	midiOutputPipeline "modularMidiGoApp/backend/midiUtility/midiOutputPipeline"

	"go.bug.st/serial"
)

// USBPortsList represents the overall structure of the JSON file.
type USBPortsList struct {
	AvailableUSBDevices []USBDevice `json:"available_usb_devices"`
	SelectedUSBDevice   string      `json:"selected_usb_device"`
}

func ESP32MidiListener(channel uint8, outputChan chan<- midiOutputPipeline.MidiCCMessage, stopChan <-chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ESP32MidiListener recovered from panic: %v", r)
		}
	}()

	for {
		select {
		case <-stopChan:
			log.Println("ESP32MidiListener stopping...")
			return
		default:
			if err := listenToESP32(channel, outputChan, stopChan); err != nil {
				log.Printf("ESP32 connection error: %v", err)
				log.Println("Retrying in 5 seconds...")

				select {
				case <-time.After(5 * time.Second):
					continue
				case <-stopChan:
					return
				}
			}
		}
	}
}

func listenToESP32(channel uint8, outputChan chan<- midiOutputPipeline.MidiCCMessage, stopChan <-chan struct{}) error {
	// Get the selected USB device
	/*deviceName, err := getSelectedUSBDevice(FilePath)
	if err != nil {
		return fmt.Errorf("failed to get USB device: %w", err)
	}
	*/
	deviceName := "/dev/ttyUSB0" // Replace with your actual device path or logic to get it dynamically
	log.Printf("Connecting to ESP32 on device: %s", deviceName)

	// Configure serial port
	mode := &serial.Mode{
		BaudRate: 115200, // Adjust baud rate as needed for your ESP32
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	// Open serial port
	port, err := serial.Open(deviceName, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}
	defer port.Close()

	log.Printf("Successfully connected to ESP32 on %s", deviceName)

	// Create buffered reader for line-by-line reading
	reader := bufio.NewReader(port)

	for {
		select {
		case <-stopChan:
			log.Println("Stopping ESP32 listener...")
			return nil
		default:
			// Read line (until newline)
			line, err := reader.ReadBytes('\n')
			if err != nil {
				// Check if it's a timeout error (normal when no data)
				if err.Error() == "timeout" {
					continue
				}
				return fmt.Errorf("failed to read from serial port: %w", err)
			}

			// Process the received data
			if err := processMidiData(line, channel, outputChan); err != nil {
				log.Printf("Error processing MIDI data: %v", err)
				continue
			}
		}
	}
}

func processMidiData(data []byte, channel uint8, outputChan chan<- midiOutputPipeline.MidiCCMessage) error {
	// Remove newline characters
	if len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r') {
		data = data[:len(data)-1]
	}
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data = data[:len(data)-1]
	}

	// Check if we have valid data (must be even number of bytes, minimum 2)
	if len(data) < 2 || len(data)%2 != 0 {
		return fmt.Errorf("invalid data length: %d bytes", len(data))
	}

	// Process pairs of bytes (CC number, value)
	for i := 0; i < len(data); i += 2 {
		ccNumber := data[i]
		value := data[i+1]

		msg := midiOutputPipeline.MidiCCMessage{
			Channel:    channel,
			Controller: uint8(ccNumber),
			Value:      uint8(value),
		}

		// Send to output channel (non-blocking)
		select {
		case outputChan <- msg:
			log.Printf("MIDI CC: Channel=%d, Controller=%d, Value=%d",
				msg.Channel, msg.Controller, msg.Value)
		default:
			log.Println("Warning: Output channel full, dropping MIDI message")
		}
	}

	return nil
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
