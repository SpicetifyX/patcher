package spotify

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func InstallStandaloneSpotify(outDir string, overwriteExisting bool) {
	if _, err := os.Stat(outDir); err == nil {
		log.Println("Spotify installation already exists in this folder")
		if overwriteExisting {
			err = os.RemoveAll(outDir)
			if err != nil {
				log.Panicln(err)
			}
		} else {
			return
		}
	}

	err := os.MkdirAll(outDir, 0700)
	if err != nil {
		log.Panicln(err)
	}

	log.Printf("Installing standalone Spotify client dir: %s\n", outDir)
	resp, err := http.Get(DIRECT_DOWNLOAD_URL)
	if err != nil {
		log.Panicln(err)
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	err = os.WriteFile(path.Join(outDir, "SpotifyFullSetup.exe"), content, 0700)
	if err != nil {
		log.Panicln(err)
	}

	ExtractSpotify(outDir)
	// os.MkdirAll(path.Join(outDir, "data", "AppData", "roaming"), 0700)
	// os.MkdirAll(path.Join(outDir, "data", "AppData", "local"), 0700)

	// cmd = exec.Command(path.Join(outDir, "Spotify.exe"))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("APPDATA=%s", path.Join(outDir, "data", "AppData", "roaming")))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("USERPROFILE=%s", path.Join(outDir, "data")))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("LOCALAPPDATA=%s", path.Join(outDir, "data", "AppData", "local")))
	// err = cmd.Run()
	// if err != nil {
	// log.Panicln(err)
	// }
}
