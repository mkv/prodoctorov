package prodoctorov_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"prodoctorov/internal/service/prodoctorov"
)

func jsonUnmarshal(t *testing.T, j string) interface{} {
	var i interface{}
	if err := json.Unmarshal([]byte(j), &i); err != nil {
		t.Errorf("failed to setup prerequisite: %v", err)
	}

	return i
}

func TestNewSchedule(t *testing.T) {
	filialSchedule, err := prodoctorov.NewSchedule("Филиал 1")
	if err != nil {
		t.Fatalf("NewSchedule() error = %v", err)
	}

	doctorSchedule, err := prodoctorov.NewDoctorSchedule("Иванов И.И.", "Аллерголог", 1)
	if err != nil {
		t.Fatalf("NewDoctorSchedule() error = %v", err)
	}

	startTime, err := time.Parse("2006-01-02T15:04:05", "2021-02-27T13:00:00")
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}

	duration, err := time.ParseDuration("30m")
	if err != nil {
		t.Fatalf("time.ParseDuration() error = %v", err)
	}

	err = doctorSchedule.AddTimeCell(startTime, duration, true, "42")
	if err != nil {
		t.Fatalf("AddTimeCell() error = %v", err)
	}

	err = filialSchedule.AddDoctorSchedule(doctorSchedule)
	if err != nil {
		t.Fatalf("AddDoctorSchedule() error = %v", err)
	}

	gotMessage, err := filialSchedule.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	wantMessage := `{"schedule":{"filial_id":"Филиал 1","data":{"filial_id":{"Аллерголог/ИвановИ.И.":{"efio":"Иванов И.И.","espec":"Аллерголог","cells":[{"dt":"2021-02-27","time_start":"13:00","time_end":"13:30","free":true,"room":"42"}]}}}}}` //nolint:revive // test data

	if !reflect.DeepEqual(jsonUnmarshal(t, string(gotMessage)), jsonUnmarshal(t, wantMessage)) {
		t.Errorf("got = %s, want %s", gotMessage, wantMessage)
	}
}
