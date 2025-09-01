package espconfigutility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	getvalues "modularMidiGoApp/backend/getValues"
	usbUtility "modularMidiGoApp/backend/usbUtility"
)

/*
specifically for the 4051BE from Texas Instruments,
most other multiplexers will also work if whired
correctly to the ESP32. Please look into the datasheets.
https://www.ti.com/lit/ds/symlink/cd4051b.pdf?HQS=dis-dk-null-digikeymode-dsf-pf-null-wwe&ts=1754731206272&ref_url=https%253A%252F%252Fwww.ti.com%252Fgeneral%252Fdocs%252Fsuppproductinfo.tsp%253FdistId%253D10%2526gotoUrl%253Dhttps%253A%252F%252Fwww.ti.com%252Flit%252Fgpn%252Fcd4051b
*/
type Multiplexer8bit struct {
	Id          int    `json:"id"`
	Pins        [8]int `json:"pins"`
	PinA        int    `json:"pinA"`
	PinB        int    `json:"pinB"`
	PinC        int    `json:"pinC"`
	SerialIOpin int    `json:"serialIOpin"`
}

func WritePinConfig() error {
	fmt.Println("Writing pin configuration...")
	mux, err := readConfFile()
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return err
	}

	fmt.Printf("Loaded %d multiplexers from config file.\n", len(mux))

	var configString bytes.Buffer
	configString.WriteString("MUX_COUNT:" + strconv.Itoa(len(mux)) + ";")

	for i, m := range mux {
		configString.WriteString(fmt.Sprintf("MUX%d:%d,%d,%d,%d,%d",
			i+1, m.Id, m.PinA, m.PinB, m.PinC, m.SerialIOpin))
		for _, pin := range m.Pins {
			configString.WriteString("," + strconv.Itoa(pin))
		}
		configString.WriteString(";")
	}

	if err := usbUtility.WriteToUSB(configString.String()); err != nil {
		fmt.Printf("Error writing to USB: %v\n", err)
		return err
	}

	fmt.Printf("Config data sent: %s\n", configString.String())
	fmt.Println("Pin configuration write completed.")
	return nil
}

type MuxConfig struct {
	Multiplexers []Multiplexer8bit `json:"multiplexers"`
}

func readConfFile() ([]Multiplexer8bit, error) {
	filePath := filepath.Join(getvalues.FindRootPath(), "/espConfigUtility/pinConf.json")
	fmt.Printf("Reading config file from: %s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening config file: %v\n", err)
		return nil, err
	}
	defer func() {
		fmt.Println("Closing config file.")
		file.Close()
	}()

	var muxConfig MuxConfig
	if err := json.NewDecoder(file).Decode(&muxConfig); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil, err
	}

	fmt.Println("Config file successfully decoded.")
	return muxConfig.Multiplexers, nil
}