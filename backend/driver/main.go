package main

import (
	"modularMidiGoApp/backend/usbUtility"
	"os"
	"path/filepath"
)

func main() {
	usbUtility.UsbPortLists(findRootPath())
}

func findRootPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	parentDir := filepath.Dir(dir)
	return parentDir
}
