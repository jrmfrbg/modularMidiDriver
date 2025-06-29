const int analogPins[5] = { 34, 35, 32, 33, 36 };  // Adjust pins as needed
const int hysteresisVal = 32;
int lastRaw = 0;
void setup() {
  Serial.begin(115200);  // Initialize serial communication
}

void loop() {
  int i = 0;
  int raw = analogRead(analogPins[i]);  // Read analog value (0-4095)
  if (raw > lastRaw && (raw - lastRaw) > hysteresisVal || lastRaw > raw && (lastRaw - raw) > hysteresisVal) {
    uint8_t mapped = map(raw, 0, 4095, 0, 127);  // Map to 7-bit range
    uint8_t currI = i;
    Serial.write(i + 1);
    Serial.write(mapped);  // Send as uint8_t
    Serial.println();
  }
  lastRaw = analogRead(analogPins[i]);
}
