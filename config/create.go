package config

import (
	"log"
	"os"
	"path"
	"strings"

	"gopkg.in/ini.v1"
)

func Create(configDir string) *Config {
	var iniData *ini.File
	var err error

	iniData, err = ini.Load(path.Join(configDir, "patcher.ini"))
	if err != nil {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			log.Panicf("Could not create config directory at: %s, err: %v\n", configDir, err)
		}

		err = os.MkdirAll(path.Join(configDir, "installations"), 0700)
		if err != nil {
			log.Panicln(err)
		}

		err = os.MkdirAll(path.Join(configDir, "extensions"), 0700)
		if err != nil {
			log.Panicln(err)
		}

		err = os.MkdirAll(path.Join(configDir, "apps"), 0700)
		if err != nil {
			log.Panicln(err)
		}

		err = os.MkdirAll(path.Join(configDir, "themes"), 0700)
		if err != nil {
			log.Panicln(err)
		}

		iniData = ini.Empty()

		section, err := iniData.NewSection("Installation")
		if err != nil {
			log.Panicln(err)
		}

		_, err = section.NewKey("current_version", "")
		if err != nil {
			log.Panicln(err)
		}

		section, err = iniData.NewSection("Spotify")
		if err != nil {
			log.Panicln(err)
		}

		_, err = section.NewKey("enable_developer_tools", "1")
		if err != nil {
			log.Panicln(err)
		}

		err = iniData.SaveTo(path.Join(configDir, "patcher.ini"))
		if err != nil {
			log.Panicln(err)
		}
	}

	section, err := iniData.GetSection("Spotify")
	if err != nil {
		log.Panicln(err)
	}

	enableDeveloperToolsStr, err := section.GetKey("enable_developer_tools")
	if err != nil {
		log.Panicln(err)
	}

	var enableDeveloperTools bool
	if strings.Contains(enableDeveloperToolsStr.String(), "1") {
		enableDeveloperTools = true
	} else {
		enableDeveloperTools = false
	}

	section, err = iniData.GetSection("Installation")
	if err != nil {
		log.Panicln(err)
	}

	currentVersion, err := section.GetKey("current_version")
	if err != nil {
		log.Panicln(err)
	}

	return &Config{EnableDeveloperTools: enableDeveloperTools, CurrentVersion: currentVersion.String(), iniData: iniData, configDir: configDir}
}
