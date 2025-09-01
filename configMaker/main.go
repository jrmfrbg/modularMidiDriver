package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// MultiplexerConfig represents the configuration for a single multiplexer.
type MultiplexerConfig struct {
	ID          int   `json:"id"`
	Pins        []int `json:"pins"`
	PinA        int   `json:"pinA"`
	PinB        int   `json:"pinB"`
	PinC        int   `json:"pinC"`
	SerialIOpin int   `json:"serialIOpin"`
}

// PinConfig represents the overall pin configuration for all multiplexers.
type PinConfig struct {
	Multiplexers []MultiplexerConfig `json:"multiplexers"`
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	muxCountStr := r.FormValue("mux-count")
	muxCount, err := strconv.Atoi(muxCountStr)
	if err != nil {
		http.Error(w, "Invalid number of multiplexers", http.StatusBadRequest)
		return
	}

	config := PinConfig{}

	for i := 0; i < muxCount; i++ {
		muxIDStr := r.FormValue("mux-" + strconv.Itoa(i) + "-id")
		muxID, _ := strconv.Atoi(muxIDStr)

		pinAStr := r.FormValue("mux-" + strconv.Itoa(i) + "-pinA")
		pinA, _ := strconv.Atoi(pinAStr)

		pinBStr := r.FormValue("mux-" + strconv.Itoa(i) + "-pinB")
		pinB, _ := strconv.Atoi(pinBStr)

		pinCStr := r.FormValue("mux-" + strconv.Itoa(i) + "-pinC")
		pinC, _ := strconv.Atoi(pinCStr)

		serialIOpinStr := r.FormValue("mux-" + strconv.Itoa(i) + "-serialIOpin")
		serialIOpin, _ := strconv.Atoi(serialIOpinStr)

		muxPins := make([]int, 8)
		for j := 0; j < 8; j++ {
			pinValueStr := r.FormValue("mux-" + strconv.Itoa(i) + "-io-" + strconv.Itoa(j))
			muxPins[j], _ = strconv.Atoi(pinValueStr)
		}

		mux := MultiplexerConfig{
			ID:          muxID,
			Pins:        muxPins,
			PinA:        pinA,
			PinB:        pinB,
			PinC:        pinC,
			SerialIOpin: serialIOpin,
		}
		config.Multiplexers = append(config.Multiplexers, mux)
	}

	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		http.Error(w, "Failed to generate JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/generate", handleGenerate)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
