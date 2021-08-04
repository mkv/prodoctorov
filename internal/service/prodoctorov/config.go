package prodoctorov

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// module errors
var (
	ErrNoURL        = errors.New("API URL not found (url option)")
	ErrBadURL       = errors.New("API URL must have trailing / (url option)")
	ErrNoFilialName = errors.New("filial name not found (filial_name option)")
	ErrNoToken      = errors.New("token not found (token option)")
)

type Config struct {
	FilialName string `yaml:"filial_name"`
	URL        string `yaml:"url"`

	Token     string `yaml:"token"`
	authToken string

	UploadDataCopyDir string `yaml:"upload_data_copy_dir"`
}

func (c *Config) Check() error {
	if c.FilialName == "" {
		return ErrNoFilialName
	}

	if c.URL == "" {
		return ErrNoURL
	}

	if !strings.HasSuffix(c.URL, "/") {
		return ErrBadURL
	}

	if c.Token == "" {
		return ErrNoToken
	}

	c.authToken = fmt.Sprintf("Token %s", c.Token)

	return nil
}

func (c *Config) isRequireRawCopy() bool {
	return c.UploadDataCopyDir != ""
}

func (c *Config) rawCopyFilename(sessionID string) string {
	return filepath.Join(c.UploadDataCopyDir, fmt.Sprintf("prodoctorov.raw.%s.json", sessionID))
}

func (c *Config) AuthToken() string {
	return c.authToken
}
