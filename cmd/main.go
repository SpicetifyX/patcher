package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"patcher/config"
	"patcher/internal/patcher"
	"patcher/internal/spotify"
)

var ConfigPath string

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Logger initialized")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Could not get user home directory: %v\n", err)
	}

	ConfigPath = path.Join(homeDir, ".patcher")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cfg := config.Create(ConfigPath)

	installDir := path.Join(ConfigPath, "installations", "dev")
	cfg.SetInstallation(installDir, "dev")

	command := os.Args[1]

	switch command {

	case "spotify":
		handleSpotify(os.Args, installDir, cfg)

	case "patch":
		handlePatch(os.Args, installDir, cfg)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func handleSpotify(args []string, installDir string, cfg *config.Config) {

	if len(args) < 3 {
		fmt.Println("Missing spotify subcommand")
		printSpotifyHelp()
		return
	}

	switch args[2] {

	case "install":
		log.Println("Installing Spotify...")
		spotify.InstallStandaloneSpotify(installDir, true)

	case "open":
		log.Println("Opening Spotify...")
		spotify.OpenSpotify(installDir)

	default:
		fmt.Printf("Unknown spotify command: %s\n", args[2])
		printSpotifyHelp()
	}
}

func handlePatch(args []string, installDir string, cfg *config.Config) {

	if len(args) < 3 {
		fmt.Println("Missing patch subcommand")
		printPatchHelp()
		return
	}

	switch args[2] {

	case "restore":
		log.Println("Restoring SPA apps...")
		patcher.RestoreSPAApps(installDir)

	case "client":
		log.Println("Patching Spotify client...")
		patcher.PatchSpotifyClient(installDir, cfg)

	default:
		fmt.Printf("Unknown patch command: %s\n", args[2])
		printPatchHelp()
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  patcher spotify install")
	fmt.Println("  patcher spotify open")
	fmt.Println("  patcher patch restore")
	fmt.Println("  patcher patch client")
}

func printSpotifyHelp() {
	fmt.Println("Spotify commands:")
	fmt.Println("  patcher spotify install")
	fmt.Println("  patcher spotify open")
}

func printPatchHelp() {
	fmt.Println("Patch commands:")
	fmt.Println("  patcher patch restore")
	fmt.Println("  patcher patch client")
}
