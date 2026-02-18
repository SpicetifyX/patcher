package spotify

import (
	"log"
	"os"
	"os/exec"
	"path"
)

func ExtractSpotify(outDir string) {
	if _, err := os.Stat(path.Join(outDir, "Apps")); err == nil {

		CleanOutDir(outDir)
	}

	cmd := exec.Command(path.Join(outDir, "SpotifyFullSetup.exe"), "/extract", outDir)
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}

}

func CleanOutDir(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		if name == "SpotifyFullSetup.exe" || name == "v8_context_snapshot.bin" {
			continue
		}

		fullPath := path.Join(outDir, name)

		err := os.RemoveAll(fullPath)
		if err != nil {
			return err
		}
	}

	return nil
}
