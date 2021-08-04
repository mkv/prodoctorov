package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"prodoctorov/internal/service/logger"
)

const (
	ModuleName = "PRODOCTOROV"

	UploadAfterStartSec = 5
)

type Service struct {
	config *Config
	log    *zap.SugaredLogger
}

func NewService(configFile string) (*Service, error) {
	cfg, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	s := Service{
		config: cfg,
	}

	return &s, nil
}

func (s *Service) Run(closeChan chan os.Signal) error {
	var err error

	ctx, ctxCancel := context.WithCancel(context.Background())

	defer ctxCancel()

	s.log, err = logger.NewLogger(ctx, ModuleName, s.config.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logging subsystem: %w", err)
	}

	ticker := time.NewTimer(UploadAfterStartSec * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-closeChan:
			s.log.Warnf("%s interrupted by signal", ModuleName)

			return nil
		case <-ticker.C:
			session, err := NewUploadSession(s.config, s.log)
			if err != nil {
				return err // fatal error
			}

			if err := session.Upload(ctx); err != nil {
				s.log.Error(err)
			}

			ticker.Reset(s.config.startEvery) // rearm timer after upload
		}
	}
}
