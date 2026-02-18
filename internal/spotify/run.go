package spotify

import (
	"log"
	"os/exec"
	"path"
)

func OpenSpotify(installDir string) {
	cmd := exec.Command(path.Join(installDir, "Spotify.exe"))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("APPDATA=%s", path.Join(installDir, "data", "AppData", "roaming")))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("USERPROFILE=%s", path.Join(installDir, "data")))
	// cmd.Env = append(cmd.Env, fmt.Sprintf("LOCALAPPDATA=%s", path.Join(installDir, "data", "AppData", "local")))
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}

}
