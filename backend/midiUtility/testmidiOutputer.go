package midiCCOutputer

import (
	"fmt"
	"log"
	"math"
	midiOutputPipeline "modularMidiGoApp/backend/midiUtility/midiOutputPipeline"
	"time"
)

// WiggleTest simulates wiggling/oscillating values for a given CC number
// ccNumber: the MIDI CC number to wiggle (0-126)
// centerValue: the center value to wiggle around (0-127)
// amplitude: how much to wiggle (0-63, will be clamped to stay within 0-127 range)
// steps: number of wiggle steps to generate
// channel: MIDI channel (0-15)
func WiggleTest(ccNumber int, centerValue int, amplitude int, steps int, channel uint8) {
	fmt.Printf("=== Wiggle Test ===\n")
	fmt.Printf("CC: %d, Center: %d, Amplitude: %d, Steps: %d, Channel: %d\n\n",
		ccNumber, centerValue, amplitude, steps, channel+1)

	if centerValue < 0 || centerValue > 127 {
		log.Printf("Center value %d out of range (0-127)", centerValue)
		return
	}
	if ccNumber < 0 || ccNumber > 126 {
		log.Printf("CC number %d out of range (0-126)", ccNumber)
		return
	}

	for i := 0; i < steps; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(steps-1)
		wiggleOffset := float64(amplitude) * math.Sin(angle)
		wiggleValue := float64(centerValue) + wiggleOffset

		if wiggleValue < 0 {
			wiggleValue = 0
		}
		if wiggleValue > 127 {
			wiggleValue = 127
		}

		// Create the specific message
		msg := midiOutputPipeline.MidiCCMessage{
			Channel:    channel,
			Controller: uint8(ccNumber),
			Value:      uint8(wiggleValue),
		}

		// Send the single, efficient message
		// NOTE: Your channel should now be of type chan MidiCCMessage
		midiOutputPipeline.MidiOutChannel <- msg

		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("Wiggle test completed\n\n")
}

// SmoothWiggleTest creates a smooth continuous wiggle effect
// ccNumber: the MIDI CC number to wiggle (0-126)
// centerValue: the center value to wiggle around (0-127)
// amplitude: how much to wiggle (0-63)
// duration: how long to wiggle in seconds
// frequency: wiggle frequency in Hz (wiggles per second)
// channel: MIDI channel (0-15)
func SmoothWiggleTest(ccNumber int, centerValue int, amplitude int, duration float64, frequency float64, channel uint8) {
	fmt.Printf("=== Smooth Wiggle Test ===\n")
	fmt.Printf("CC: %d, Center: %d, Amplitude: %d, Duration: %.1fs, Frequency: %.1fHz, Channel: %d\n\n",
		ccNumber, centerValue, amplitude, duration, frequency, channel+1)

	if centerValue < 0 || centerValue > 127 {
		log.Printf("Center value %d out of range (0-127)", centerValue)
		return
	}
	if ccNumber < 0 || ccNumber > 126 {
		log.Printf("CC number %d out of range (0-126)", ccNumber)
		return
	}

	startTime := time.Now()

	for time.Since(startTime).Seconds() < duration {
		// Calculate current time position
		elapsed := time.Since(startTime).Seconds()

		// Create sine wave based on frequency
		angle := elapsed * frequency * 2.0 * math.Pi

		// Calculate wiggle value
		wiggleOffset := float64(amplitude) * math.Sin(angle)
		wiggleValue := float64(centerValue) + wiggleOffset

		if wiggleValue < 0 {
			wiggleValue = 0
		}
		if wiggleValue > 127 {
			wiggleValue = 127
		}

		// Create the specific message
		msg := midiOutputPipeline.MidiCCMessage{
			Channel:    channel,
			Controller: uint8(ccNumber),
			Value:      uint8(wiggleValue),
		}

		// Send the single, efficient message
		// NOTE: Your channel should now be of type chan MidiCCMessage
		midiOutputPipeline.MidiOutChannel <- msg
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("Smooth wiggle test completed\n\n")
}

// RandomWiggleTest creates random wiggle values around a center point
// ccNumber: the MIDI CC number to wiggle (0-126)
// centerValue: the center value to wiggle around (0-127)
// maxDeviation: maximum random deviation from center (0-63)
// count: number of random values to send
// delay: delay between messages in milliseconds
// channel: MIDI channel (0-15)
func RandomWiggleTest(ccNumber int, centerValue int, maxDeviation int, count int, delay int, channel uint8) {
	fmt.Printf("=== Random Wiggle Test ===\n")
	fmt.Printf("CC: %d, Center: %d, Max Deviation: %d, Count: %d, Delay: %dms, Channel: %d\n\n",
		ccNumber, centerValue, maxDeviation, count, delay, channel+1)

	if centerValue < 0 || centerValue > 127 {
		log.Printf("Center value %d out of range (0-127)", centerValue)
		return
	}
	if ccNumber < 0 || ccNumber > 126 {
		log.Printf("CC number %d out of range (0-126)", ccNumber)
		return
	}

	for i := 0; i < count; i++ {
		// Generate random deviation
		deviation := (math.Sin(float64(i)*0.3) * float64(maxDeviation)) +
			(math.Cos(float64(i)*0.7) * float64(maxDeviation) * 0.5)

		wiggleValue := float64(centerValue) + deviation

		if wiggleValue < 0 {
			wiggleValue = 0
		}
		if wiggleValue > 127 {
			wiggleValue = 127
		}

		// Create the specific message
		msg := midiOutputPipeline.MidiCCMessage{
			Channel:    channel,
			Controller: uint8(ccNumber),
			Value:      uint8(wiggleValue),
		}

		// Send the single, efficient message
		// NOTE: Your channel should now be of type chan MidiCCMessage
		midiOutputPipeline.MidiOutChannel <- msg

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	fmt.Printf("Random wiggle test completed\n\n")
}

// Update StartTest to use new function signatures
func StartTest() {
	// Test wiggling CC 1 (Modulation) around value 64 with amplitude 30
	WiggleTest(1, 64, 30, 20, 0)

	// Test smooth wiggling CC 7 (Volume) around 100 for 3 seconds at 2Hz
	SmoothWiggleTest(7, 100, 20, 3.0, 2.0, 0)

	// Test random wiggling CC 10 (Pan) around center with max deviation 40
	RandomWiggleTest(10, 64, 40, 15, 100, 0)

	// Test wiggling CC 74 (Filter Cutoff) - common for synths
	fmt.Println("Testing filter cutoff wiggle...")
	SmoothWiggleTest(74, 80, 25, 2.0, 1, 0)
}
