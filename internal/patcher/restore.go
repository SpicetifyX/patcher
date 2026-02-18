package patcher

import (
	"log"
	"patcher/internal/spotify"
)

func RestoreSPAApps(clientDir string) {
	log.Println("Restoring SPA apps...")
	spotify.ExtractSpotify(clientDir)
}
