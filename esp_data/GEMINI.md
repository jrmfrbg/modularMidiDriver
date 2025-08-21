# Plan for ESP32 modularMidiDriver Firmware

## 1. Project Goals

The primary goal of this firmware is to create a modular and extensible system for the ESP32 to act as a MIDI controller. The firmware will be responsible for reading various hardware inputs (potentiometers, buttons, etc.), translating them into MIDI messages, and sending them to the backend application running on a PC via a serial (USB) connection.

## 2. Core Functionality

### Hardware Input
- Read analog values from potentiometers using the ESP32's ADC.
- Read digital values from buttons and switches.
- Implement a mechanism to detect significant changes in sensor values to avoid flooding the backend with redundant data.

### MIDI Message Generation
- Map raw sensor values to the standard MIDI range (0-127).
- Create MIDI Control Change (CC) messages. Each message will consist of a channel, a CC number (controller), and a value.

### Serial Communication
- Establish a serial communication link with the PC backend over USB.
- The data will be sent in a format that the backend's `processMidiData` function can parse. Based on the backend code, this format is a stream of byte pairs `(controller, value)` followed by a newline character.

## 3. Code Structure

The Arduino `.ino` file will be structured as follows:

### Includes
- `Arduino.h`: Core Arduino library.

### Configuration
- **Pin Definitions**: Use `#define` or `const int` to declare the ESP32 pins connected to sensors.
- **MIDI Mapping**: Create a data structure (e.g., an array of structs) to map each input pin to a specific MIDI CC number. This will make the code easy to configure and expand.
- **State Management**: An array or struct to hold the last known state of each sensor to track changes.

### `setup()` function
- Initialize the serial communication with the same baud rate as the backend (e.g., 9600).
- Configure the `pinMode` for each sensor pin (e.g., `INPUT` for potentiometers, `INPUT_PULLUP` for buttons).

### `loop()` function
- The main loop will continuously call a function to read all sensors.
- It will check if a sensor's value has changed more than a certain threshold since the last reading.
- If there's a change, it will format the data as a MIDI CC message and send it over the serial port.

### Helper Functions
- `readSensors()`: A function that iterates through the configured sensors, reads their current values, and stores them.
- `sendMidiData(uint8_t controller, uint8_t value)`: A function that takes a controller and value, formats them into a two-byte array, and sends them over serial, followed by a newline.

## 4. Communication Protocol with Backend

The communication protocol will adhere to the format expected by the `processMidiData` function in the Go backend.

- **Data Packet**: A sequence of one or more MIDI messages.
- **Message Format**: Each message is a pair of bytes: `[controller_number, value]`.
- **Packet Termination**: The sequence of byte pairs is terminated by a newline character (`\n`).

**Example**: If a potentiometer mapped to CC 20 changes to a value of 100, the ESP32 will send the bytes `0x14` (20) and `0x64` (100) followed by a newline.

## 5. Example Implementation Details

### Reading a Potentiometer
```cpp
int potValue = analogRead(POT_PIN);
int midiValue = map(potValue, 0, 4095, 0, 127); // ESP32 has a 12-bit ADC
```

### Handling a Button Press
```cpp
int buttonState = digitalRead(BUTTON_PIN);
if (buttonState != lastButtonState) {
  if (buttonState == LOW) { // Assuming INPUT_PULLUP
    sendMidiData(BUTTON_CC, 127); // Note On
  } else {
    sendMidiData(BUTTON_CC, 0);   // Note Off
  }
  lastButtonState = buttonState;
}
```

## 6. Future Development (Roadmap)

- **Dynamic Configuration**: Implement a mechanism to receive configuration data from the backend over serial. This would allow for remapping controls without reflashing the ESP32.
- **Support for More Hardware**: Add support for I2C and SPI to connect a wider range of modules and sensors (e.g., I/O expanders, encoders).
- **UDP Communication**: Implement wireless communication over WiFi using UDP, as envisioned in the project's `README.md`.
- **Heartbeat/Status Messages**: Periodically send a status message to the backend to indicate that the device is connected and running.
