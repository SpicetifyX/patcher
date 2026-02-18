package patcher

import (
	"log"
	"os"
	"patcher/config"
	"patcher/patches"
	"path"
)

func PatchSpotifyClient(clientDir string, config *config.Config) {
	log.Printf("Patching spotify client located at: %s\n", clientDir)
	ExtractSPAApps(clientDir)
	if err := patches.PatchDevTools(path.Join(os.Getenv("LOCALAPPDATA"), "Spotify", "offline.bnk"), config.EnableDeveloperTools); err != nil {
		log.Printf("Could not toggle devtools, %v\n", err)
	}

	if err := patches.PatchV8ContextSnapshot(path.Join(clientDir, "v8_context_snapshot.bin"), path.Join(clientDir, "Apps", "xpui")); err != nil {
		log.Panicln(err)
	}

	patches.StartMinimal(path.Join(clientDir, "Apps"))
	patches.AdditionalOptions(path.Join(clientDir, "Apps"))
}
