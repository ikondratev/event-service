package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	settingsPath = "settings"
	fileName = "settings.json"
)

type Settings struct {
	Environment string `json:"environment"`
	Server 		Server  `json:"server"`
	Db	   		Db	   `json:"db"`
	Kafka		Kafka  `json:"kafka"`
	Worker 		Worker `json:"worker"`
}

type Server struct {
	Port            string `json:"port"`
	WaitingShutdown int    `json:"waiting_shutdown"`
	HeaderTimeout   int    `json:"header_timout"`
	ReadTimeout     int    `json:"read_timout"`
	WriteTimeout    int    `json:"write_timeout"`
	IdleTimeout     int    `json:"idle_timeout"` 
}

type Kafka struct {
	Brokers 	   []string 		 `json:"brokers"`
	Topics  	   map[string]string `json:"topics"`
	PollInterval   int 				 `json:"poll_interval"`
	BatchSize 	   int 				 `json:"batch_size"`
	FlushTimeout   int 				 `json:"flush_timeout"`
}

type Db struct {
	Url 		   string `json:"url"`
	OpenConnection    int `json:"open_conn"`
	MaxIdleConnection int `json:"max_idle_conn"`
	MaxLifeTime		  int `json:"max_life_time"`
	StartDelay		  int `json:"start_delay"`
}

type Worker struct {
	LeaseSeconds 	int `json:"lease_seconds"`
	MaxAttempts 	int `json:"max_attempts"`
	BackoffSeconds 	int `json:"backoff_seconds"`
	TimeoutSeconds	int `json:"timeout_seconds"`
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

	if path := os.Getenv("DB_URL_FILE"); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("db url file: %w", err)
		}
		set.Db.Url = strings.TrimSpace(string(raw))
	}

	return &set, nil
}