package domino

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"prodoctorov/internal/service/dominocsv"
)

// constants
const (
	MinFieldsCount = 7

	BusyCode = "busy"

	IdxSpec      = 0
	IdxName      = 1
	IdxStartTime = 2
	IdxDuration  = 3
	IdxFree      = 4
	IdxRoom      = 5

	TimeLayout = "2.1.06 15:04:05"

	DefaultMeetDuration = 20 * time.Minute
	MaxMeetDuration     = 120 * time.Minute
)

// errors
var (
	ErrMalformedRecord = errors.New("malformed record")
	ErrExpiredRecord   = errors.New("expired record")
	ErrMandatoryField  = errors.New("mandatory field is empty")
)

type Record struct {
	Spec      string
	Name      string
	StartTime time.Time
	Duration  time.Duration
	Free      bool
	Room      string
}

func (r *Record) ID() string {
	return fmt.Sprintf("%s/%s", r.Spec, r.Name)
}

type Records []*Record

type OrderByID Records

func (a OrderByID) Len() int      { return len(a) }
func (a OrderByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderByID) Less(i, j int) bool {
	return a[i].ID() < a[j].ID()
}

type OrderByDate Records

func (a OrderByDate) Len() int      { return len(a) }
func (a OrderByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderByDate) Less(i, j int) bool {
	return a[i].StartTime.Before(a[j].StartTime)
}

func NewRecord(record []string, timeNow time.Time) (*Record, error) {
	if len(record) < MinFieldsCount {
		return nil, fmt.Errorf("%w: %s", ErrMalformedRecord, "too short record")
	}

	result := &Record{
		Spec: record[IdxSpec],
		Name: record[IdxName],
		Free: record[IdxFree] != BusyCode, // empty field equals free time
		Room: record[IdxRoom],
	}

	if result.Spec == "" {
		return nil, fmt.Errorf("%w: %s", ErrMandatoryField, "spec")
	}

	if result.Name == "" {
		return nil, fmt.Errorf("%w: %s", ErrMandatoryField, "name")
	}

	var err error

	result.StartTime, err = time.Parse(TimeLayout, record[IdxStartTime])
	if err != nil {
		return nil, fmt.Errorf("failed to decode starting date field: %w", err)
	}

	duration := record[IdxDuration]
	if duration != "" {
		result.Duration, err = time.ParseDuration(fmt.Sprintf("%sm", duration))
		if err != nil {
			return nil, fmt.Errorf("failed to decode interval field: %w", err)
		}
	}

	monthBegin := time.Date(timeNow.Year(), timeNow.Month(), 1, 0, 0, 0, 0, time.UTC)

	if result.StartTime.Before(monthBegin) {
		return nil, ErrExpiredRecord
	}

	return result, nil
}

type LogMalformedRecord func(string)

func ImportRecords(body []byte, timeNow time.Time, log LogMalformedRecord) (Records, error) {
	csv, err := dominocsv.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	records := make(CsvRecords, 0)

	for {
		record, err := csv.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	records = records[1:] // skip header

	result := make(Records, len(records))

	i := 0

	for _, r := range records {
		rec, err := NewRecord(r, timeNow)
		if err != nil {
			if !errors.Is(err, ErrExpiredRecord) {
				log(fmt.Sprintf("skip malformed record: %v: %v", err, r))
			}

			continue
		}

		result[i] = rec
		i++
	}

	result = result[:i] // fix size, some records might be skipped

	sort.Sort(OrderByID(result))

	return result, nil
}

func equalDay(d1 time.Time, d2 time.Time) bool {
	d1y, d1m, d1d := d1.Date()
	d2y, d2m, d2d := d2.Date()

	if d1d != d2d || d1m != d2m || d1y != d2y {
		return false
	}

	return true
}

// Cleaned function fixes empty meeting duration time
func (records Records) Cleaned() Records {
	recordsCount := len(records)
	if recordsCount == 0 {
		return records
	}

	if recordsCount == 1 {
		r := records[0]
		if r.Duration == 0 {
			r.Duration = DefaultMeetDuration
		}

		return records
	}

	sort.Sort(OrderByDate(records))

	return fixDuration(recordsCount, records)
}

func fixDuration(recordsCount int, records Records) Records {
	result := make(Records, recordsCount)

	var (
		lastDate       time.Time
		actualDuration time.Duration
	)

	for i, r := range records {
		if !equalDay(lastDate, r.StartTime) {
			actualDuration = 0 // the next day there may be a different time for the duration of the meetings
		}

		nextIndex := i + 1

		if actualDuration == 0 && nextIndex < recordsCount {
			firstTime := records[i].StartTime
			secondTime := records[nextIndex].StartTime

			if secondTime.After(firstTime) && equalDay(firstTime, secondTime) {
				actualDuration = secondTime.Sub(firstTime)
			}
		}

		if r.Duration == 0 {
			if actualDuration == 0 {
				r.Duration = DefaultMeetDuration
			} else {
				r.Duration = actualDuration
			}
		}

		if r.Duration > MaxMeetDuration {
			r.Duration = DefaultMeetDuration
		}

		lastDate = r.StartTime
		result[i] = r
	}

	return result
}

type TimeCell struct {
	StartTime time.Time
	Duration  time.Duration
	Free      bool
	Room      string
}

type TimeCells []*TimeCell

type DoctorSchedule struct {
	Spec  string
	Name  string
	Cells TimeCells
}

type DoctorScheduleFetcher func(*DoctorSchedule)

// LoadDoctorSchedule external function is called every time one doctor's schedule is completed and ready to be uploaded
func (records Records) LoadDoctorSchedule(fetcher DoctorScheduleFetcher) {
	var lastID, lastSpec, lastName string

	doctorRecords := make(Records, 0)

	for _, r := range records {
		if lastID == "" {
			lastID = r.ID()
			lastSpec = r.Spec
			lastName = r.Name
		}

		if r.ID() != lastID {
			doctorRecords = doctorRecords.Cleaned()

			schedule := &DoctorSchedule{
				Spec:  lastSpec,
				Name:  lastName,
				Cells: make(TimeCells, len(doctorRecords)),
			}

			for i, rr := range doctorRecords {
				schedule.Cells[i] = &TimeCell{
					StartTime: rr.StartTime,
					Duration:  rr.Duration,
					Free:      rr.Free,
					Room:      rr.Room,
				}
			}

			fetcher(schedule)

			lastID = r.ID()
			lastSpec = r.Spec
			lastName = r.Name
			doctorRecords = make(Records, 0)
		}

		doctorRecords = append(doctorRecords, r)
	}
}
