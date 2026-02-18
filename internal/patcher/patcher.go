package patcher

import (
	"log"
)

func PatchSpotifyClient(clientDir string) {
	log.Printf("Patching spotify client located at: %s\n", clientDir)
	ExtractSPAApps(clientDir)
}
