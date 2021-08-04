package prodoctorov

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrStartAfterEnd = errors.New("wrong meeting time, start after end")
)

const singleFilial = "filial_id"

type timeCellDto struct {
	Date      string `json:"dt"`
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
	Free      bool   `json:"free"`
	Room      string `json:"room"`
}

type doctorScheduleDto struct {
	Name  string        `json:"efio"`
	Spec  string        `json:"espec"`
	Cells []timeCellDto `json:"cells"`
}

type filialID string
type doctorID string

type doctorScheduleMap map[doctorID]doctorScheduleDto

type filialScheduleMap map[filialID]doctorScheduleMap

type scheduleDto struct {
	FilialID string            `json:"filial_id"`
	Data     filialScheduleMap `json:"data"`
}

type Schedule struct {
	schedule scheduleDto
}

func NewSchedule(filialID string) (*Schedule, error) {
	s := &Schedule{
		schedule: scheduleDto{
			FilialID: filialID,
			Data:     filialScheduleMap{singleFilial: make(doctorScheduleMap, 0)},
		},
	}

	return s, nil
}

func (s *Schedule) ToJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schedule interface{} `json:"schedule"`
	}{s.schedule})
}

func (s *Schedule) AddDoctorSchedule(doc *DoctorSchedule) error {
	s.schedule.Data[singleFilial][doc.doctorID()] = doc.schedule

	return nil
}

func (s *Schedule) IsEmpty() bool {
	return len(s.schedule.Data[singleFilial]) == 0
}

type DoctorSchedule struct {
	schedule doctorScheduleDto
}

func NewDoctorSchedule(name string, spec string, cellsCount int) (*DoctorSchedule, error) {
	d := &DoctorSchedule{
		schedule: doctorScheduleDto{
			Name:  name,
			Spec:  spec,
			Cells: make([]timeCellDto, 0, cellsCount),
		},
	}

	return d, nil
}

func (d *DoctorSchedule) doctorID() doctorID {
	return doctorID(fmt.Sprintf("%s/%s", d.schedule.Spec, strings.ReplaceAll(d.schedule.Name, " ", "")))
}

func (d *DoctorSchedule) AddTimeCell(startTime time.Time, duration time.Duration, free bool, room string) error {
	cell := timeCellDto{
		Date:      startTime.Format("2006-01-02"),
		TimeStart: startTime.Format("15:04"),
		TimeEnd:   startTime.Add(duration).Format("15:04"),
		Free:      free,
		Room:      room,
	}

	if cell.TimeStart >= cell.TimeEnd {
		return ErrStartAfterEnd
	}

	d.schedule.Cells = append(d.schedule.Cells, cell)

	return nil
}
