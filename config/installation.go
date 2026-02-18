package config

import (
	"log"
	"path"
	"strings"
)

func (c *Config) GetSpotifyInstallations() {
}

func (c *Config) SetInstallation(dir string, version string) {
	if version == "" {
		log.Println("Version is not defined, adding anyway")
	}

	var cleanedPath string
	if strings.Contains(c.configDir, dir) {
		cleanedPath = strings.Trim(c.configDir, dir)
	} else {
		cleanedPath = dir
	}

	log.Printf("Setting current Spotify installation to: %s", cleanedPath)

	section, err := c.iniData.GetSection("Installation")
	if err != nil {
		log.Panicln(err)
	}

	key, err := section.GetKey("current_version")
	if err != nil {
		log.Panicln(err)
	}

	key.SetValue(cleanedPath)
	c.iniData.SaveTo(path.Join(c.configDir, "patcher.ini"))
}
