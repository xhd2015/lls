package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const FileName = "config.json"

type Config struct {
	OpenCmd string `json:"open_cmd"`

	HistoryFiles []string `json:"history_files"`

	Envs []string `json:"envs"`

	Projects []string `json:"projects"`
}

func DefaultFile(createDir bool) (string, error) {
	conf, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(conf, "lls")
	if createDir {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(configDir, FileName), nil
}

func Load(file string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("reading config %s: %w", file, err)
	}
	return cfg, nil
}

func LoadDefault(createDir bool) (Config, string, error) {
	file, err := DefaultFile(createDir)
	if err != nil {
		return Config{}, "", err
	}
	cfg, err := Load(file)
	if err != nil {
		return Config{}, file, err
	}
	return cfg, file, nil
}

func ExpandPath(path string, envs []string) string {
	const homePrefix = "~/"
	if suffix, ok := strings.CutPrefix(path, homePrefix); ok {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home + "/" + suffix
		}
	}
	return os.Expand(path, func(name string) string {
		for _, env := range envs {
			if strings.HasPrefix(env, name+"=") {
				return env[len(name)+1:]
			}
		}
		return os.Getenv(name)
	})
}

func CollapsePath(path string, envs []string) string {
	short := path
	for _, env := range envs {
		var envName string
		var envValue string
		idx := strings.Index(env, "=")
		if idx < 0 {
			envName = env
			envValue = os.Getenv(envName)
		} else {
			envName = env[:idx]
			envValue = env[idx+1:]
		}
		if envName == "" || envValue == "" {
			continue
		}
		if suffix, ok := strings.CutPrefix(short, envValue+"/"); ok {
			short = "$" + envName + "/" + suffix
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		if suffix, ok := strings.CutPrefix(short, home+"/"); ok {
			short = "~/" + suffix
		}
	}

	return short
}
