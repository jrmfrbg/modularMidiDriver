package usbUtility

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

// SerialManager handles USB serial communication with ESP32
type SerialManager struct {
	port           serial.Port
	portName       string
	isRunning      bool
	stopChan       chan struct{}
	initString     string
	triggerString  string
	responseString string
}

// NewSerialManager creates a new SerialManager instance
func NewSerialManager(portName, initString, triggerString, responseString string) *SerialManager {
	return &SerialManager{
		portName:       portName,
		stopChan:       make(chan struct{}),
		initString:     initString,
		triggerString:  triggerString,
		responseString: responseString,
	}
}

// Start begins the serial communication as a goroutine
func (sm *SerialManager) Start() error {
	// Configure serial port
	mode := &serial.Mode{
		BaudRate: 115200, // Common ESP32 baud rate
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	// Open the serial port
	port, err := serial.Open(sm.portName, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %v", sm.portName, err)
	}

	sm.port = port
	sm.isRunning = true

	// Start the communication goroutine
	go sm.run()

	return nil
}

// Stop gracefully stops the serial communication
func (sm *SerialManager) Stop() {
	if !sm.isRunning {
		return
	}

	log.Println("Stopping serial communication...")
	close(sm.stopChan)

	// Wait a moment for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	if sm.port != nil {
		sm.port.Close()
	}

	sm.isRunning = false
	log.Println("Serial communication stopped")
}

// IsRunning returns the current status of the serial manager
func (sm *SerialManager) IsRunning() bool {
	return sm.isRunning
}

// run is the main goroutine that handles serial communication
func (sm *SerialManager) run() {
	defer func() {
		if sm.port != nil {
			sm.port.Close()
		}
		sm.isRunning = false
	}()

	log.Printf("Starting serial communication on port %s", sm.portName)

	// Send initial string
	if sm.initString != "" {
		err := sm.writeString(sm.initString)
		if err != nil {
			log.Printf("Failed to send initial string: %v", err)
			return
		}
		log.Printf("Sent initial string: %s", sm.initString)
	}

	// Create a buffered reader for incoming data
	reader := bufio.NewReader(sm.port)

	for {
		select {
		case <-sm.stopChan:
			log.Println("Received stop signal")
			return

		default:
			// Set a short read timeout to allow checking stop channel
			sm.port.SetReadTimeout(100 * time.Millisecond)

			// Read incoming data
			data, err := reader.ReadString('\n')
			if err != nil {
				// Check if it's a timeout error (expected)
				if strings.Contains(err.Error(), "timeout") {
					continue
				}
				log.Printf("Error reading from serial port: %v", err)
				continue
			}

			// Clean up the received string
			receivedString := strings.TrimSpace(data)
			if receivedString == "" {
				continue
			}

			log.Printf("Received: %s", receivedString)

			// Check if received string matches trigger string
			if receivedString == sm.triggerString {
				log.Printf("Trigger string matched, sending response: %s", sm.responseString)

				err := sm.writeString(sm.responseString)
				if err != nil {
					log.Printf("Failed to send response string: %v", err)
				}
			}
		}
	}
}

// writeString writes a string to the serial port with newline
func (sm *SerialManager) writeString(data string) error {
	if sm.port == nil {
		return fmt.Errorf("serial port not open")
	}

	// Add newline if not present
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	_, err := sm.port.Write([]byte(data))
	return err
}

// SendString sends a custom string to the ESP32 (can be used externally)
func (sm *SerialManager) SendString(data string) error {
	if !sm.isRunning {
		return fmt.Errorf("serial manager is not running")
	}
	return sm.writeString(data)
}

// Example usage function demonstrating how to use the module
func ExampleUsage() {
	// Create serial manager with specific strings
	serialMgr := NewSerialManager(
		"/dev/ttyUSB0",  // Port name (Linux/Mac) - use "COM3" etc. for Windows
		"HELLO_ESP32",   // Initial string to send
		"REQUEST_DATA",  // String to listen for
		"DATA_RESPONSE", // String to send back when trigger is received
	)

	// Start the serial communication
	err := serialMgr.Start()
	if err != nil {
		log.Fatalf("Failed to start serial communication: %v", err)
	}

	// Simulate running for some time
	log.Println("Serial communication running...")

	// In your actual application, you would check for WiFi connection here
	// and call serialMgr.Stop() when WiFi is available

	// For demonstration, stop after 30 seconds
	time.Sleep(30 * time.Second)

	// Stop the serial communication (call this when WiFi is connected)
	serialMgr.Stop()
}
