package settings

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	settingsPath = "settings"
	fileName = "settings.json"
)


type Settings struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	Environment     string `json:"environment"`
	WaitingShutdown int    `json:"waiting_shutdown"`
}

func New(env string) (*Settings, error) {
	var set Settings
	filePath := fmt.Sprintf("%s/%s.%s", settingsPath, env, fileName )
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("LoadFileError: %w", err)
	}

	if err := json.Unmarshal(data, &set); err != nil {
		return nil, fmt.Errorf("UnmarshalFileError: %w", err)
	}

	return &set, nil
}