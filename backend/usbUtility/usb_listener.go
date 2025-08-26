package usbUtility

import (
	"bufio"
	"fmt"
	"log"
	midiOutputPipeline "modularMidiGoApp/backend/midiUtility/midiOutputPipeline"
	"time"

	"go.bug.st/serial"
)

// USBPortsList represents the overall structure of the JSON file.
type USBPortsList struct {
	AvailableUSBDevices []USBDevice `json:"available_usb_devices"`
	SelectedUSBDevice   string      `json:"selected_usb_device"`
}

var stopChan = make(chan struct{})

func StopESP32MidiListener() {
	close(stopChan)
}

func ESP32MidiListener(channel uint8, outputChan chan<- midiOutputPipeline.MidiCCMessage) {
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
			if err := listenToESP32(channel, outputChan); err != nil {
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

func listenToESP32(channel uint8, outputChan chan<- midiOutputPipeline.MidiCCMessage) error {
	// Get the selected USB device
	deviceName, err := GetSelectedUSBDevice(FilePath)
	if err != nil {
		return fmt.Errorf("failed to get USB device: %w", err)
	}
	log.Printf("Connecting to ESP32 on device: %s", deviceName)

	// Configure serial port
	mode := &serial.Mode{
		BaudRate: 9600, // Adjust baud rate as needed for your ESP32
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

	// Wait for ESP32 to be ready for mode selection
	modeSelectReader := bufio.NewReader(port)
	port.SetReadTimeout(30 * time.Second)
	for {
		line, err := modeSelectReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from USB device: %w", err)
		}
		if line == "SELECT_MODE\n" || line == "SELECT_MODE\r\n" {
			log.Println("ESP32 ready for mode selection")
			break
		}
	}
	//send mode selection command
	if _, err := port.Write([]byte("BROADCAST_MODE\n")); err != nil {
		return fmt.Errorf("failed to write mode selection to USB device: %w", err)
	}

	// Create buffered reader for line-by-line reading
	reader := bufio.NewReader(port)

	for {
		select {
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

func WriteToUSB(data interface{}) error {
	// Get the selected USB device
	deviceName, err := GetSelectedUSBDevice(FilePath)
	if err != nil {
		return fmt.Errorf("failed to get USB device: %w", err)
	}

	// Open a connection to the USB device
	conn, err := serial.Open(deviceName, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return fmt.Errorf("failed to open USB connection: %w", err)
	}
	defer conn.Close()

	// Convert data to []byte if necessary
	var bytesToWrite []byte
	switch v := data.(type) {
	case []byte:
		bytesToWrite = v
	case string:
		bytesToWrite = []byte(v)
	default:
		return fmt.Errorf("unsupported data type for writing to USB device")
	}

	//Wait for ESP32 to be ready for mode selection
	reader := bufio.NewReader(conn)
	conn.SetReadTimeout(30 * time.Second)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from USB device: %w", err)
		}
		if line == "SELECT_MODE\n" || line == "SELECT_MODE\r\n" {
			log.Println("ESP32 ready for mode selection")
			break
		}
	}

	//send mode selection command
	if _, err := conn.Write([]byte("FLASH_MODE\n")); err != nil {
		return fmt.Errorf("failed to write mode selection to USB device: %w", err)
	}

	// Write the data to the USB device
	if _, err := conn.Write(bytesToWrite); err != nil {
		return fmt.Errorf("failed to write to USB device: %w", err)
	}

	return nil
}
