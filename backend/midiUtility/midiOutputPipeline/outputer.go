package midioutputpipeline

import (
	"encoding/json"
	"fmt"
	"log"
	getvalues "modularMidiGoApp/backend/getValues"
	"os"
	"path/filepath"
	"regexp"
	"unicode/utf8"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

type SelectedPortStruct struct {
	SelectedPort SelectedPortData `json:"selected_midi_port"`
}

type SelectedPortData struct {
	Name     string `json:"name"`
	PortPath string `json:"port_path"`
}

var (
	rootPath_MO   = getvalues.FindRootPath()
	dirPath_MO    = filepath.Join(rootPath_MO, "midiUtility")
	midi_filePath = filepath.Join(dirPath_MO, "midi_ports.json")
)

var MidiOutChannel = make(chan [127]uint8)

func MidiWriter() {
	// Get available MIDI outputs
	defer midi.CloseDriver()

	outs := midi.GetOutPorts()
	if len(outs) == 0 {
		fmt.Println("No MIDI output ports available")
		return
	}

	// Set the selected port
	portIdx, err := getSelectedPort()
	if err != nil {
		fmt.Printf("Error getting selected port: %v\n", err)
		return
	}
	outPortID := outs[portIdx]
	send, err := midi.SendTo(outPortID)

	outChannel := MidiOutChannel

	for msg := range outChannel {
		fmt.Printf("Received MIDI message: %v\n", msg)
		for i, value := range msg {
			err := send(midi.ControlChange(0, uint8(i), value))
			if err != nil {
				log.Printf("Error sending CC %d with value %d: %v", i, value, err)
				continue
			}
		}
	}
}

func getPortPathFromOuts(inputStr string) string {
	// Ensure we handle strings shorter than 5 characters gracefully.
	if utf8.RuneCountInString(inputStr) < 5 {
		// If the string is shorter than 5 runes, process the whole string.
		re := regexp.MustCompile(`\s+`)
		return re.ReplaceAllString(inputStr, "")
	}

	// Get the last 5 runes (characters) from the string.
	// This is important for UTF-8 safety.
	runes := []rune(inputStr)
	lastFiveRunes := runes[len(runes)-5:]
	lastFiveStr := string(lastFiveRunes)

	// Compile a regular expression to find all whitespace characters.
	// \s matches space, tab, newline, and other unicode spaces.
	// + matches one or more occurrences.
	re := regexp.MustCompile(`\s+`)

	// Replace all found whitespace with an empty string.
	outputStr := re.ReplaceAllString(lastFiveStr, "")

	return outputStr
}

func getSelectedPort() (int, error) {
	outs := midi.GetOutPorts()
	if len(outs) == 0 {
		return 0, fmt.Errorf("no MIDI output ports available")
	}
	portPath, err := getMIDIFileContent()
	if err != nil {
		return 0, fmt.Errorf("failed to get MIDI file content: %v", err)
	}
	for i, out := range outs {
		fmt.Printf("Checking port %d: %s\n", i+1, out.String()) // Debug output
		fmt.Printf("Expected port path: %s\n", portPath)        // Debug output
		if getPortPathFromOuts(out.String()) == portPath {
			fmt.Printf("Selected port found: %s\n", out.String()) // Debug output
			return i, nil
		}
	}
	return 0, fmt.Errorf("selected MIDI port not found: %s", portPath)

}

func getMIDIFileContent() (string, error) {
	midi_filePath := filepath.Join(dirPath_MO, "midi_ports.json")

	fileContent, err := os.ReadFile(midi_filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read MIDI file: %v", err)
	}
	var selectedPort SelectedPortStruct
	if err := json.Unmarshal(fileContent, &selectedPort); err != nil {
		return "", fmt.Errorf("failed to parse MIDI file JSON: %v", err)
	}

	// Return the port_path instead of the whole object
	return selectedPort.SelectedPort.PortPath, nil
}
