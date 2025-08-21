package espconfigutility

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	getvalues "modularMidiGoApp/backend/getValues"
	//usbUtility "modularMidiGoApp/backend/usbUtility"
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
	mux, err := readConfFile()
	if err != nil {
		return err
	}

	// Marshal the data and write it to the USB
	for _, m := range mux {
		jsonData, err := json.Marshal(m)
		if err != nil {
			return err
		}
		/*
			if err := usbUtility.WriteToUSB(jsonData); err != nil {
				return err
			}
		*/
		fmt.Println(string(jsonData))
	}

	return nil
}

func readConfFile() ([]Multiplexer8bit, error) {
	filePath := filepath.Join(getvalues.FindRootPath(), "/backend/espConfigUtility/pinConf.json")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mux []Multiplexer8bit
	if err := json.NewDecoder(file).Decode(&mux); err != nil {
		return nil, err
	}

	return mux, nil
}
