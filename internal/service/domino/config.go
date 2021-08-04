package domino

import "errors"

var (
	ErrNoURL = errors.New("URL not found (url option)")
)

type Config struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	RawScheduleCopyDir string `yaml:"raw_schedule_copy_dir"`
}

func (c *Config) Check() error {
	if c.URL == "" {
		return ErrNoURL
	}

	return nil
}
