#include <Arduino.h>
#include <Preferences.h>

// --- Configuration ---

// Mode selector
enum Mode {
  BROADCAST,
  FLASH
};
Mode currentMode = BROADCAST;

// Pin Definitions (Example, please configure for your hardware)
#define MUX_COUNT 5
#define MUX_IO_PIN_BASE 1 // Example, will be set from config

// MIDI Mapping
struct Mux {
  int id;
  int pins[8];
  int pinA;
  int pinB;
  int pinC;
  int serialIOpin;
};

Mux multiplexers[MUX_COUNT];
int lastSensorValues[MUX_COUNT][8];

// Threshold for detecting significant change
#define CHANGE_THRESHOLD 4

// --- Preferences for Flash Storage ---
Preferences preferences;

// --- Helper Functions ---

void sendMidiData(uint8_t controller, uint8_t value) {
  Serial.write(controller);
  Serial.write(value);
  Serial.write('\n');
}

void readSensors() {
  for (int muxIndex = 0; muxIndex < MUX_COUNT; muxIndex++) {
    for (int pinIndex = 0; pinIndex < 8; pinIndex++) {
      // Select the pin on the multiplexer
      digitalWrite(multiplexers[muxIndex].pinA, (pinIndex & 1) ? HIGH : LOW);
      digitalWrite(multiplexers[muxIndex].pinB, (pinIndex & 2) ? HIGH : LOW);
      digitalWrite(multiplexers[muxIndex].pinC, (pinIndex & 4) ? HIGH : LOW);

      // Read the value
      int sensorValue = analogRead(multiplexers[muxIndex].serialIOpin);
      int midiValue = map(sensorValue, 0, 4095, 0, 127);

      // Check for significant change
      if (abs(midiValue - lastSensorValues[muxIndex][pinIndex]) > CHANGE_THRESHOLD) {
        lastSensorValues[muxIndex][pinIndex] = midiValue;
        int controller = multiplexers[muxIndex].pins[pinIndex];
        sendMidiData(controller, midiValue);
      }
    }
  }
}

void loadConfigFromFlash() {
  preferences.begin("midiConfig", false);
  String configString = preferences.getString("config", "");
  // TODO: Parse the configString and populate the multiplexers array.
  // The config string should contain the number of multiplexers,
  // and for each multiplexer, its ID, pinA, pinB, pinC, serialIOpin,
  // and the 8 MIDI CC numbers for its pins.
  // Example format: "MUX_COUNT:5;MUX1:id,pA,pB,pC,io,c1,c2,c3,c4,c5,c6,c7,c8;MUX2:..."
  preferences.end();
}

void saveConfigToFlash(String config) {
  preferences.begin("midiConfig", false);
  preferences.putString("config", config);
  preferences.end();
}

// --- Main Functions ---

void setup() {
  Serial.begin(9600);

  // Load configuration from flash memory
  loadConfigFromFlash();

  // After loading, configure the pins for each multiplexer
  for (int i = 0; i < MUX_COUNT; i++) {
    pinMode(multiplexers[i].pinA, OUTPUT);
    pinMode(multiplexers[i].pinB, OUTPUT);
    pinMode(multiplexers[i].pinC, OUTPUT);
    pinMode(multiplexers[i].serialIOpin, INPUT);
  }

  Serial.println("SELECT_MODE");
}

void loop() {
  // Check for incoming serial data for mode switching
  if (Serial.available() > 0) {
    String input = Serial.readStringUntil('\n');
    if (input == "FLASH_MODE") {
      currentMode = FLASH;
    } else if (input == "BROADCAST_MODE") {
      currentMode = BROADCAST;
      // After switching to broadcast mode, reload the config and re-init pins
      loadConfigFromFlash();
      for (int i = 0; i < MUX_COUNT; i++) {
        pinMode(multiplexers[i].pinA, OUTPUT);
        pinMode(multiplexers[i].pinB, OUTPUT);
        pinMode(multiplexers[i].pinC, OUTPUT);
        pinMode(multiplexers[i].serialIOpin, INPUT);
      }
    } else if (currentMode == FLASH) {
      saveConfigToFlash(input);
      Serial.println("Config saved!");
    }
  }

  if (currentMode == BROADCAST) {
    readSensors();
  }
  // In FLASH mode, we just wait for config data
}