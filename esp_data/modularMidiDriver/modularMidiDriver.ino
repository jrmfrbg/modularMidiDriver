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

#define TX2 11
#define RX2 10
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

bool validateConfig(String configString) {
  if (configString == "") {
    return false;
  }

  // Parse MUX_COUNT
  int muxCountIndex = configString.indexOf("MUX_COUNT:");
  if (muxCountIndex == -1) {
    return false;
  }
  int muxCount = configString.substring(muxCountIndex + 10).toInt();
  if (muxCount > MUX_COUNT) {
    muxCount = MUX_COUNT;
  }

  // Parse each multiplexer's data
  int startIndex = 0;
  for (int i = 0; i < muxCount; i++) {
    String muxKey = "MUX" + String(i + 1) + ":";
    int muxIndex = configString.indexOf(muxKey, startIndex);
    if (muxIndex == -1) {
      return false;
    }
    int endIndex = configString.indexOf(';', muxIndex);
    if (endIndex == -1) {
      endIndex = configString.length();
    }
    String muxData = configString.substring(muxIndex + muxKey.length(), endIndex);
    
    // Split the muxData by commas
    int valueIndex = 0;
    int lastIndex = 0;
    String values[13];
    for(int j = 0; j < 13; j++) {
        int currentIndex = muxData.indexOf(',', lastIndex);
        if (currentIndex == -1) {
            currentIndex = muxData.length();
        }
        values[j] = muxData.substring(lastIndex, currentIndex);
        lastIndex = currentIndex + 1;
    }

    if (lastIndex < muxData.length()) {
        return false;
    }
  }
  return true;
}

void loadConfigFromFlash() {
  preferences.begin("midiConfig", false);
  String configString = preferences.getString("config", "");
  if (!validateConfig(configString)) {
    Serial.println("Invalid config found in flash.");
    return;
  }

  // Parse MUX_COUNT
  int muxCountIndex = configString.indexOf("MUX_COUNT:");
  int muxCount = configString.substring(muxCountIndex + 10).toInt();
  if (muxCount > MUX_COUNT) {
    muxCount = MUX_COUNT;
  }

  // Parse each multiplexer's data
  int startIndex = 0;
  for (int i = 0; i < muxCount; i++) {
    String muxKey = "MUX" + String(i + 1) + ":";
    int muxIndex = configString.indexOf(muxKey, startIndex);
    int endIndex = configString.indexOf(';', muxIndex);
    if (endIndex == -1) {
      endIndex = configString.length();
    }
    String muxData = configString.substring(muxIndex + muxKey.length(), endIndex);
    
    // Split the muxData by commas
    int valueIndex = 0;
    int lastIndex = 0;
    String values[13];
    for(int j = 0; j < 13; j++) {
        int currentIndex = muxData.indexOf(',', lastIndex);
        if (currentIndex == -1) {
            currentIndex = muxData.length();
        }
        values[j] = muxData.substring(lastIndex, currentIndex);
        lastIndex = currentIndex + 1;
    }

    multiplexers[i].id = values[0].toInt();
    multiplexers[i].pinA = values[1].toInt();
    multiplexers[i].pinB = values[2].toInt();
    multiplexers[i].pinC = values[3].toInt();
    multiplexers[i].serialIOpin = values[4].toInt();
    for (int j = 0; j < 8; j++) {
      multiplexers[i].pins[j] = values[j + 5].toInt();
    }
    startIndex = endIndex;
  }
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
  sleep(1);
  Serial.println("STARTING UP! Serial1");
  Serial2.begin(9600, SERIAL_8N1, RX2, TX2);
  sleep(1);
  Serial2.println("STARTING UP! Serial2");
  sleep(1);
  Serial.println("STARTED UP!");
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
      if (validateConfig(input)) {
        saveConfigToFlash(input);
        Serial.println("Config saved!");
      } else {
        Serial.println("Invalid config format!");
      }
    } else {
      Serial.println("Unknown command in BROADCAST mode.");
    }
  }
}