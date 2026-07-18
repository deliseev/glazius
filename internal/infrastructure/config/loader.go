package config

import (
	"encoding/json"
	"os"

	"github.com/deliseev/glazius/internal/domain/entity"
)

func Load(path string) (*entity.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg entity.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
