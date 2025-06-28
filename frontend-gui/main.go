package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
)

type USBDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"device_path"`
}

type USBDeviceData struct {
	AvailableUSBDevices []USBDevice `json:"available_usb_devices"`
	SelectedUSBDevice   string      `json:"selected_usb_device"`
}

type MIDIDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"device_path"`
}

type MIDIDeviceData struct {
	AvailableMIDIDevices []USBDevice `json:"available_midi_devices"`
	SelectedMIDIDevice   string      `json:"selected_midi_device"`
}

type DeviceManager struct {
	rootPath           string
	confPath           string
	backendApiLocation string

	// UI elements
	usbList     *widget.List
	midiList    *widget.List
	statusLabel *widget.Label
	refreshBtn  *widget.Button
	testMidiBtn *widget.Button

	// Data
	usbData      *USBDeviceData
	midiData     *MIDIDeviceData
	usbFilePath  string
	midiFilePath string
}

func NewDeviceManager() *DeviceManager {
	dm := &DeviceManager{
		rootPath: getRootPath(),
	}
	dm.confPath = filepath.Join(dm.rootPath, "backend", "modularMidi.conf")
	dm.backendApiLocation = dm.generateBackendApiLocation()
	return dm
}

func (dm *DeviceManager) generateBackendApiLocation() string {
	httpconfRAW := dm.loadHTTPconf()
	return strings.Join([]string{parseProtocol(httpconfRAW), "://", parseHost(httpconfRAW), ":", parsePort(httpconfRAW)}, "")
}

func (dm *DeviceManager) loadHTTPconf() string {
	cfg, err := ini.Load(dm.confPath)
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

	return strings.Join([]string{
		"listen_port:",
		getKey("http", "listen_port"),
		",backend_api_port:",
		getKey("http", "backend_api_port"),
		",backend_api_host:",
		getKey("http", "backend_api_host"),
		",backend_api_protocol:",
		getKey("http", "backend_api_protocol"),
	}, "")
}

func (dm *DeviceManager) getUSBFileContent() error {
	resp, err := http.Get(strings.Join([]string{dm.backendApiLocation, "/usbPortListFile"}, ""))
	if err != nil {
		return fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %v", err)
	}

	filePath := strings.TrimSpace(string(body))
	if filePath == "" {
		return fmt.Errorf("API returned empty file path")
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var usbData USBDeviceData
	err = json.Unmarshal(fileContent, &usbData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	dm.usbData = &usbData
	dm.usbFilePath = filePath
	return nil
}

func (dm *DeviceManager) getMIDIFileContent() error {
	resp, err := http.Get(strings.Join([]string{dm.backendApiLocation, "/listMidiPorts"}, ""))
	if err != nil {
		return fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %v", err)
	}

	filePath := strings.TrimSpace(string(body))
	if filePath == "" {
		return fmt.Errorf("API returned empty file path")
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var midiData MIDIDeviceData
	err = json.Unmarshal(fileContent, &midiData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	dm.midiData = &midiData
	dm.midiFilePath = filePath
	return nil
}

func (dm *DeviceManager) selectUSBDevice(index int) error {
	if dm.usbData == nil || index < 0 || index >= len(dm.usbData.AvailableUSBDevices) {
		return fmt.Errorf("invalid USB device index")
	}

	selectedDevice := dm.usbData.AvailableUSBDevices[index]
	dm.usbData.SelectedUSBDevice = selectedDevice.DevicePath

	updatedJSON, err := json.MarshalIndent(dm.usbData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(dm.usbFilePath, updatedJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

func (dm *DeviceManager) selectMIDIDevice(index int) error {
	if dm.midiData == nil || index < 0 || index >= len(dm.midiData.AvailableMIDIDevices) {
		return fmt.Errorf("invalid MIDI device index")
	}

	selectedDevice := dm.midiData.AvailableMIDIDevices[index]
	dm.midiData.SelectedMIDIDevice = selectedDevice.DevicePath

	updatedJSON, err := json.MarshalIndent(dm.midiData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(dm.midiFilePath, updatedJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

func (dm *DeviceManager) testMidiOutput() error {
	resp, err := http.Get(strings.Join([]string{dm.backendApiLocation, "/testMidiOutput"}, ""))
	if err != nil {
		return fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (dm *DeviceManager) updateStatus(message string) {
	dm.statusLabel.SetText(fmt.Sprintf("Status: %s", message))
}

func (dm *DeviceManager) refreshData(window fyne.Window) {
	dm.updateStatus("Refreshing devices...")

	go func() {
		// Refresh USB devices
		err := dm.getUSBFileContent()
		if err != nil {
			dm.updateStatus(fmt.Sprintf("Error loading USB devices: %v", err))
		}

		// Refresh MIDI devices
		err = dm.getMIDIFileContent()
		if err != nil {
			dm.updateStatus(fmt.Sprintf("Error loading MIDI devices: %v", err))
		}

		// Update UI on main thread
		dm.usbList.Refresh()
		dm.midiList.Refresh()
		dm.updateStatus("Devices refreshed successfully")
	}()
}

func (dm *DeviceManager) createUSBList() *widget.List {
	list := widget.NewList(
		func() int {
			if dm.usbData == nil {
				return 0
			}
			return len(dm.usbData.AvailableUSBDevices)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.ComputerIcon()),
				widget.NewLabel("Template"),
				layout.NewSpacer(),
				widget.NewLabel("Selected"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if dm.usbData == nil || id >= len(dm.usbData.AvailableUSBDevices) {
				return
			}

			device := dm.usbData.AvailableUSBDevices[id]
			container := obj.(*fyne.Container)

			// Update device name
			label := container.Objects[1].(*widget.Label)
			label.SetText(device.Name)

			// Update selection indicator
			selectedLabel := container.Objects[3].(*widget.Label)
			if device.DevicePath == dm.usbData.SelectedUSBDevice {
				selectedLabel.SetText("✓ Selected")
				selectedLabel.Importance = widget.HighImportance
			} else {
				selectedLabel.SetText("")
				selectedLabel.Importance = widget.MediumImportance
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		err := dm.selectUSBDevice(id)
		if err != nil {
			dm.updateStatus(fmt.Sprintf("Error selecting USB device: %v", err))
		} else {
			device := dm.usbData.AvailableUSBDevices[id]
			dm.updateStatus(fmt.Sprintf("Selected USB device: %s", device.Name))
			list.Refresh()
		}
	}

	return list
}

func (dm *DeviceManager) createMIDIList() *widget.List {
	list := widget.NewList(
		func() int {
			if dm.midiData == nil {
				return 0
			}
			return len(dm.midiData.AvailableMIDIDevices)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.MediaMusicIcon()),
				widget.NewLabel("Template"),
				layout.NewSpacer(),
				widget.NewLabel("Selected"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if dm.midiData == nil || id >= len(dm.midiData.AvailableMIDIDevices) {
				return
			}

			device := dm.midiData.AvailableMIDIDevices[id]
			container := obj.(*fyne.Container)

			// Update device name
			label := container.Objects[1].(*widget.Label)
			label.SetText(device.Name)

			// Update selection indicator
			selectedLabel := container.Objects[3].(*widget.Label)
			if device.DevicePath == dm.midiData.SelectedMIDIDevice {
				selectedLabel.SetText("✓ Selected")
				selectedLabel.Importance = widget.HighImportance
			} else {
				selectedLabel.SetText("")
				selectedLabel.Importance = widget.MediumImportance
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		err := dm.selectMIDIDevice(id)
		if err != nil {
			dm.updateStatus(fmt.Sprintf("Error selecting MIDI device: %v", err))
		} else {
			device := dm.midiData.AvailableMIDIDevices[id]
			dm.updateStatus(fmt.Sprintf("Selected MIDI device: %s", device.Name))
			list.Refresh()
		}
	}

	return list
}

func (dm *DeviceManager) createMainWindow() fyne.Window {
	myApp := app.New()
	myApp.SetIcon(theme.ComputerIcon())

	window := myApp.NewWindow("Modular MIDI Driver by Jorim")
	window.SetIcon(theme.ComputerIcon())
	window.Resize(fyne.NewSize(800, 1000))

	// Create UI elements
	dm.statusLabel = widget.NewLabel("Status: Ready")
	dm.statusLabel.Importance = widget.MediumImportance

	dm.refreshBtn = widget.NewButtonWithIcon("Refresh Devices", theme.ViewRefreshIcon(), func() {
		dm.refreshData(window)
	})

	dm.testMidiBtn = widget.NewButtonWithIcon("Test MIDI Output", theme.MediaPlayIcon(), func() {
		err := dm.testMidiOutput()
		if err != nil {
			dm.updateStatus(fmt.Sprintf("MIDI test failed: %v", err))
			dialog.ShowError(err, window)
		} else {
			dm.updateStatus("MIDI test successful")
			dialog.ShowInformation("MIDI Test", "MIDI output test completed successfully!", window)
		}
	})

	// Create device lists
	dm.usbList = dm.createUSBList()
	dm.midiList = dm.createMIDIList()

	// Create layout
	usbCard := widget.NewCard("USB Devices", "Select a USB device from the list below",
		container.NewBorder(nil, nil, nil, nil, dm.usbList))

	midiCard := widget.NewCard("MIDI Devices", "Select a MIDI device from the list below",
		container.NewBorder(nil, nil, nil, nil, dm.midiList))

	buttonContainer := container.NewHBox(
		dm.refreshBtn,
		dm.testMidiBtn,
	)

	mainContent := container.NewVBox(
		container.NewHBox(usbCard, midiCard),
		container.NewBorder(nil, nil, nil, nil, buttonContainer),
		dm.statusLabel,
	)

	window.SetContent(mainContent)

	// Load initial data
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to ensure UI is ready
		dm.refreshData(window)
	}()

	return window
}

// Helper functions (unchanged from original)
func getRootPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	parentDir := filepath.Dir(dir)
	return parentDir
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
	return protocol
}

func main() {
	dm := NewDeviceManager()
	window := dm.createMainWindow()
	window.ShowAndRun()
}
