package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"prodoctorov/internal/service/domino"
	"prodoctorov/internal/service/prodoctorov"
)

type CsvRecords [][]string

type UploadSession struct {
	config    *Config
	sessionID string

	log *zap.SugaredLogger
}

func NewUploadSession(config *Config, parentLogger *zap.SugaredLogger) (*UploadSession, error) {
	sessionID := sessionID()

	e := &UploadSession{
		config:    config,
		sessionID: sessionID,
		log:       parentLogger.Named(fmt.Sprintf("UPLOAD:%s", sessionID)),
	}

	return e, nil
}

func sessionID() string {
	return time.Now().Format("20060102T150405.999999999")
}

func (s *UploadSession) Upload(ctx context.Context) error {
	s.log.Info("Start schedule upload")

	defer s.log.Info("Schedule upload done")

	dominoSchedule, err := domino.DownloadSchedule(
		ctx,
		&s.config.Domino,
		s.sessionID,
		func(message string) {
			s.log.Error(message)
		},
	)
	if err != nil {
		return err
	}

	schedule, err := CreateSchedule(
		s.config.Prodoctorov.FilialName,
		dominoSchedule.Schedule(),
		func(message string) {
			s.log.Error(message)
		},
	)
	if err != nil {
		return err
	}

	return prodoctorov.Upload(
		ctx,
		&s.config.Prodoctorov,
		s.sessionID,
		func(message string) {
			s.log.Error(message)
		},
		schedule,
	)
}

type ErrorLogger func(string)

func CreateSchedule(filialID string, dominoSchedule domino.Records, log ErrorLogger) (*prodoctorov.Schedule, error) {
	schedule, err := prodoctorov.NewSchedule(filialID)
	if err != nil {
		return nil, err
	}

	dominoSchedule.LoadDoctorSchedule(func(export *domino.DoctorSchedule) {
		doctorSchedule, err := prodoctorov.NewDoctorSchedule(export.Name, export.Spec, len(export.Cells))
		if err != nil {
			log(fmt.Sprintf("failed to create a new doctors schedule: %v: %v", err, export))

			return
		}

		for _, cell := range export.Cells {
			if err := doctorSchedule.AddTimeCell(cell.StartTime, cell.Duration, cell.Free, cell.Room); err != nil {
				log(fmt.Sprintf("failed to append time cell: %v: %s/%s: %v", err, export.Spec, export.Name, cell))

				continue
			}
		}

		if err := schedule.AddDoctorSchedule(doctorSchedule); err != nil {
			log(fmt.Sprintf("failed to append a doctor schedule: %v", err))

			return
		}
	})

	return schedule, err
}
