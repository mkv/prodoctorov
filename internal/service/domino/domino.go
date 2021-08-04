package domino

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"
)

const (
	MaxResponseBodySize = 100 * 1024 * 1024
	DownloadTimeout     = 60 * time.Second
)

var (
	ErrDownloadFailed = errors.New("failed to download schedule")
)

// CsvRecords type for Domino export representation
type CsvRecords [][]string

type Domino struct {
	config    *Config
	sessionID string
	records   Records
}

type LogError func(string)

func DownloadSchedule(ctx context.Context, config *Config, sessionID string, log LogError) (*Domino, error) {
	d := &Domino{
		config:    config,
		sessionID: sessionID,
	}

	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.URL, nil)
	if err != nil {
		return nil, err
	}

	if config.Username != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log(err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code: %v", ErrDownloadFailed, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(&io.LimitedReader{R: resp.Body, N: MaxResponseBodySize})
	if err != nil {
		return nil, err
	}

	if d.isRequireDominoRawCopy() {
		if err := ioutil.WriteFile(d.dominoRawCopyFilename(), body, 0600); err != nil {
			log(err.Error())
		}
	}

	d.records, err = ImportRecords(body, time.Now(), func(message string) {
		log(message)
	})
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Domino) isRequireDominoRawCopy() bool {
	return d.config.RawScheduleCopyDir != ""
}

func (d *Domino) dominoRawCopyFilename() string {
	return filepath.Join(d.config.RawScheduleCopyDir, fmt.Sprintf("domino.raw.%s.csv", d.sessionID))
}

func (d *Domino) Schedule() Records {
	return d.records
}
