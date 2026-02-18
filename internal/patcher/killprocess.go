package patcher

import "os/exec"

func KillSpotifyProcess() error {
	cmd := exec.Command("taskkill", "/IM", "Spotify.exe", "/F")
	return cmd.Run()
}
