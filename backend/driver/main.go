package main

import (
	"log"
	httphandler "modularMidiGoApp/backend/httpHandler"
	midiOutputPipeline "modularMidiGoApp/backend/midiUtility/midiOutputPipeline"
	usbUtility "modularMidiGoApp/backend/usbUtility"
	"strings"
)

// Executes first and prepares:
// - Starts HTTP handler
func main() {
	go midiOutputPipeline.MidiWriter()
	go usbUtility.USBListener()

	go func() {
		routes := []httphandler.Route{
			httphandler.TestCallRoute,
			httphandler.UsbPortList,
			httphandler.MidiTester,
			httphandler.MidiPortList,
			// Add more routes
		}
		port := parsePort(LoadHTTPconf())
		if err := httphandler.StartHTTPServer(port, routes); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	log.Println("HTTP handler started successfully.")

	// Continue with other async tasks or main logic here
	log.Println("running async tasks...")

	// Prevent main from exiting immediately (example: block forever)
	select {}
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
