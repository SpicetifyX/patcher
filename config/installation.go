package config

import (
	"log"
	"strings"
)

func (c *Config) GetSpotifyInstallations() {
}

func (c *Config) SetInstallation(path string, version string) {
	if version == "" {
		log.Println("Version is not defined, adding anyway")
	}

	cleanedPath := strings.Trim(c.configDir, path)
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
}
