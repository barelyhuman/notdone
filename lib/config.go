package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port string `json:"port"`
}

var config = &Config{
	Port: "3000",
}

func LoadConfig() *Config {
	defaultConfigJSON, err := json.Marshal(config)
	Bail(err)

	baseDir := GetBaseDirectory()
	configPath := filepath.Join(baseDir, "config.json")
	_, err = os.Stat(configPath)

	if os.IsNotExist(err) {
		os.WriteFile(configPath, defaultConfigJSON, os.ModePerm)
	} else {
		Bail(err)
	}

	_config, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	json.Unmarshal(_config, &config)

	return config
}
