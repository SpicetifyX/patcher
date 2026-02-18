package main

import (
	"log"
	"os"
	"patcher/config"
	"path"
)

var ConfigPath string

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Logger initialized")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Could not get user home directory, %v", err)
	}

	ConfigPath = path.Join(homeDir, ".patcher")
}

func main() {
	config := config.Create(ConfigPath)

	log.Println("Loaded config")
	log.Printf("[config] developer_tools_enabled: %v", config.EnableDeveloperTools)
	log.Printf("[config] current_version: %v", config.CurrentVersion)
}
