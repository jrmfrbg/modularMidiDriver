package handlers

import (
	"encoding/json"
	"modularMidiGoApp/backend/usbUtility"
	"net/http"
	//"modularMidiGoApp/backend/driver"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := usbUtility.UsbPortLists()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
