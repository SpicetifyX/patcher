package main

import (
	"log"
	"os"
	"patcher/config"
	"patcher/internal/patcher"
	"patcher/internal/spotify"
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

	config.SetInstallation(path.Join(ConfigPath, "installations", "dev"), "dev")
	spotify.InstallStandaloneSpotify(path.Join(ConfigPath, "installations", "dev"), true)
	// patcher.RestoreSPAApps(path.Join(ConfigPath, "installations", "dev"))
	patcher.PatchSpotifyClient(path.Join(ConfigPath, "installations", "dev"), config)
	// spotify.OpenSpotify(path.Join(ConfigPath, "installations", "dev"))
}
