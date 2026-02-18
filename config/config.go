package config

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	EnableDeveloperTools bool
	CurrentVersion       string
	iniData              *ini.File
	configDir            string
}
