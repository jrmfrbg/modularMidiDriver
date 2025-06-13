package main

import (
	"log"
	getvalues "modularMidiGoApp/backend/getValues"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	rootPath  string = getvalues.FindRootPath()                    // Gets root path from main.go
	confPath  string = filepath.Join(rootPath, "modularMidi.conf") // Edited rootPath to lead to .conf File
	returnStr string                                               // String to be modifyed by functions
)

func LoadHTTPconf() string {
	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	getKey := func(section, key string) string {
		s := cfg.Section(section)
		if !s.HasKey(key) {
			log.Fatalf("Missing key [%s] %s", section, key)
		}
		return s.Key(key).String()
	}

	returnStr := strings.Join([]string{
		"listen_port:",
		getKey("http", "listen_port"),
		",backend_api_port:",
		getKey("http", "backend_api_port"),
		",backend_api_host:",
		getKey("http", "backend_api_host"),
		",backend_api_protocol:",
		getKey("http", "backend_api_protocol"),
		";",
	}, "")
	return returnStr
}

func LoadUDPconf() string {
	cfg, err := ini.Load(confPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	getKey := func(section, key string) string {
		s := cfg.Section(section)
		if !s.HasKey(key) {
			log.Fatalf("Missing key [%s] %s", section, key)
		}
		return s.Key(key).String()
	}

	returnStr := strings.Join([]string{
		"listen_port:",
		getKey("udp", "listen_port"),
		",send_port:",
		getKey("udp", "send_port"),
		",backend_host:",
		getKey("udp", "backend_host"),
		";",
	}, "")
	return returnStr
}
