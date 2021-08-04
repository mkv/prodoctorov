package prodoctorov

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrBadStatusCode = errors.New("unexpected HTTP status code")
)

const (
	MaxResponseBodySize = 1024 * 1024
	UploadTimeout       = 60 * time.Second
)

type LogError func(string)

func Upload(ctx context.Context, config *Config, sessionID string, log LogError, schedule *Schedule) error {
	if schedule.IsEmpty() {
		log("Schedule data to upload is empty, nothing to do")

		return nil
	}

	scheduleData, err := schedule.ToJSON()
	if err != nil {
		return err
	}

	if config.isRequireRawCopy() {
		if err := ioutil.WriteFile(config.rawCopyFilename(sessionID), scheduleData, 0600); err != nil {
			log(err.Error())
		}
	}

	ctx, cancel := context.WithTimeout(ctx, UploadTimeout)

	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewBuffer(scheduleData))
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", config.AuthToken())
	request.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log(err.Error())
		}
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, err := ioutil.ReadAll(&io.LimitedReader{R: response.Body, N: MaxResponseBodySize})
		if err != nil {
			log(err.Error()) // may be useful error message
		}

		return fmt.Errorf("%w (%v): body: %s", ErrBadStatusCode, response.StatusCode, body)
	}

	return nil
}
