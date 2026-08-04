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
	Host        string `json:"host"`
	Port        string `json:"port"`
	Environment string `json:"environment"`
}

func New(env string) *Settings  {
	var set Settings
	filePath := fmt.Sprintf("%s/%s.%s", settingsPath, env, fileName )
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &set); err != nil {
		panic(err)
	}

	return &set
}