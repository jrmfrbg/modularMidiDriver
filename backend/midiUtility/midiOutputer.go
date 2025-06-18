package midiCCOutputer

import (
	"fmt"
	"log"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

// OutputMIDICC sends MIDI Control Change messages based on the provided matrix
// matrix should be a 2D slice where:
// - matrix[0] contains CC numbers (0-127)
// - matrix[1] contains CC values (0-127)
// Both rows should have the same length
func OutputMIDICC(matrix [][]int, channel uint8) error {
	// Validate input
	if len(matrix) != 2 {
		return fmt.Errorf("matrix must have exactly 2 rows (CC numbers and values)")
	}
	if len(matrix[0]) != len(matrix[1]) {
		return fmt.Errorf("CC numbers and values rows must have the same length")
	}
	if channel > 15 {
		return fmt.Errorf("MIDI channel must be 0-15, got %d", channel)
	}

	// Get available MIDI outputs
	defer midi.CloseDriver()

	outs := midi.GetOutPorts()
	if len(outs) == 0 {
		return fmt.Errorf("no MIDI output ports available")
	}

	// Use the first available output port
	out := outs[0]

	// Open the MIDI output port
	send, err := midi.SendTo(out)
	if err != nil {
		return fmt.Errorf("failed to open MIDI output: %v", err)
	}

	fmt.Printf("Sending MIDI CC messages to: %s\n", out.String())

	// Send CC messages
	for i := 0; i < len(matrix[0]); i++ {
		ccNum := matrix[0][i]
		ccVal := matrix[1][i]

		// Validate CC number and value ranges
		if ccNum < 0 || ccNum > 127 {
			log.Printf("Warning: CC number %d is out of range (0-127), skipping", ccNum)
			continue
		}
		if ccVal < 0 || ccVal > 127 {
			log.Printf("Warning: CC value %d is out of range (0-127), skipping", ccVal)
			continue
		}

		// Send the CC message
		err := send(midi.ControlChange(channel, uint8(ccNum), uint8(ccVal)))
		if err != nil {
			log.Printf("Error sending CC %d with value %d: %v", ccNum, ccVal, err)
			continue
		}

		fmt.Printf("Sent CC %d = %d on channel %d\n", ccNum, ccVal, channel+1)
	}

	return nil
}
