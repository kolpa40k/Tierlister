package storage

import (
	"errors"
	"os"
	"path/filepath"
)

const settingsFile = "settings.json"
const defaultTheme = "Классическая"

type Settings struct {
	ThemeName string `json:"theme_name"`
}

func settingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tierlist-app", settingsFile), nil
}

func LoadSettings() (*Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return nil, err
	}
	var settings Settings
	if err := readJSON(path, &settings); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if settings.ThemeName == "" {
		settings.ThemeName = defaultTheme
	}
	return &settings, nil
}

func SaveSettings(settings *Settings) error {
	if settings == nil {
		return errors.New("settings must not be nil")
	}
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return writeJSONAtomic(path, settings)
}
