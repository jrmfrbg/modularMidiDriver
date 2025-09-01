package httphandler

import (
	"fmt"
	espConfigUtility "modularMidiGoApp/backend/espConfigUtility"
	midiCCOutputer "modularMidiGoApp/backend/midiUtility"
	midiOutputPipeline "modularMidiGoApp/backend/midiUtility/midiOutputPipeline"
	"modularMidiGoApp/backend/usbUtility"
	"net/http"
)

// Route defines a mapping between a URL path and its handler function.
type Route struct {
	Path    string
	Handler http.HandlerFunc
}

var StartNormalMode = Route{
	Path: "/startNormalMode",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		go usbUtility.ESP32MidiListener(0, midiOutputPipeline.MidiOutChannel)
		fmt.Fprint(w, "Normal mode started")
	},
}

var TestCallRoute = Route{
	Path: "/testCall",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello Client")
	},
}

var WritePinConfig = Route{
	Path: "/writePinConfig",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		espConfigUtility.WritePinConfig()
		fmt.Fprint(w, "Pin configuration written successfully.")
	},
}

var UsbPortList = Route{
	Path: "/usbPortListFile",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		usb_ports_file := usbUtility.UsbPortLists()
		fmt.Fprint(w, usb_ports_file)
	},
}

var MidiTester = Route{
	Path: "/testMidiOutput",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		go midiCCOutputer.StartTest()
		fmt.Fprint(w, "Midi Output Test Triggered")
	},
}

var MidiPortList = Route{
	Path: "/listMidiPorts",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		midi_ports_file := midiCCOutputer.ListMIDIPorts()
		fmt.Fprintln(w, midi_ports_file)
	},
}

var StartUSBListener = Route{
	Path: "/startUSBListener",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		go usbUtility.ESP32MidiListener(0, midiOutputPipeline.MidiOutChannel)
		fmt.Fprint(w, "USB listener started successfully.")
	},
}

// Package httphandler provides functionality to start an HTTP server with specific routes

// StartHTTPServer starts an HTTP server on the given port and uses the provided routes for GET requests.
func StartHTTPServer(port string, routes []Route) error {
	mux := http.NewServeMux()
	for _, route := range routes {
		// Wrap each handler to only allow GET requests
		mux.HandleFunc(route.Path, func(handler http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					handler(w, r)
				} else {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				}
			}
		}(route.Handler))
	}
	addr := fmt.Sprintf(":%s", port)
	return http.ListenAndServe(addr, mux)
}
