package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"prodoctorov/internal/service/domino"
	"prodoctorov/internal/service/prodoctorov"
)

var (
	ErrConfigNotFound = errors.New("configuration file not found")
)

// defaults
const (
	DefaultLogLevel = "info"

	DefaultStartEveryMinutes = 60
)

// Config root service configuration
type Config struct {
	LogLevel string `yaml:"log_level"`

	Domino domino.Config `yaml:"domino"`

	Prodoctorov prodoctorov.Config `yaml:"prodoctorov"`

	StartEveryMinutes int `yaml:"start_every_minutes"`
	startEvery        time.Duration
}

func configExists(fileName string) bool {
	info, err := os.Stat(filepath.Clean(fileName))
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// LoadConfig reads config file in the YAML format
func LoadConfig(fileName string) (*Config, error) {
	log.Printf("Loading config file '%s'...", fileName)

	if !configExists(fileName) {
		return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, fileName)
	}

	yamlCfg, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	err = yaml.Unmarshal(yamlCfg, cfg)
	if err != nil {
		return nil, err
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = DefaultLogLevel
	}

	if cfg.StartEveryMinutes <= 0 {
		cfg.StartEveryMinutes = DefaultStartEveryMinutes
	}

	cfg.startEvery = time.Duration(cfg.StartEveryMinutes) * time.Minute

	if err := cfg.Domino.Check(); err != nil {
		return nil, fmt.Errorf("bad domino config: %w", err)
	}

	if err := cfg.Prodoctorov.Check(); err != nil {
		return nil, fmt.Errorf("bad prodoctorov config: %w", err)
	}

	return cfg, nil
}
