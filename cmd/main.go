package main

import (
	"log"
	"os"
	"patcher/config"
	"patcher/internal/patcher"
	"path"
)

var ConfigPath string

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Logger initialized")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Could not get user home directory, %v\n", err)
	}

	ConfigPath = path.Join(homeDir, ".patcher")
}

func main() {
	config := config.Create(ConfigPath)

	log.Println("Loaded config")
	log.Printf("[config] developer_tools_enabled: %v\n", config.EnableDeveloperTools)
	log.Printf("[config] current_version: %v\n", config.CurrentVersion)

	// spotify.InstallStandaloneSpotify(path.Join(ConfigPath, "installations", "dev"), true)
	patcher.PatchSpotifyClient(path.Join(ConfigPath, "installations", "dev"))
	// patcher.RestoreSPAApps(path.Join(ConfigPath, "installations", "dev"))
	// spotify.OpenSpotify(path.Join(ConfigPath, "installations", "dev"))
}
