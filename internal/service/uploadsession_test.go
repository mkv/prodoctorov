package service_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"prodoctorov/internal/service"
	"prodoctorov/internal/service/domino"
)

var (
	//go:embed testdata.csv
	fromCSV []byte

	//go:embed testdata.json
	toJSON []byte
)

func jsonUnmarshal(t *testing.T, j []byte) interface{} {
	var i interface{}
	if err := json.Unmarshal(j, &i); err != nil {
		t.Errorf("failed to unmarshal JSON: %v", err)
	}

	return i
}

func TestCreateSchedule(t *testing.T) {
	type args struct {
		filialID       string
		dominoSchedule domino.Records
		log            service.ErrorLogger
	}

	dominoSchedule, err := domino.ImportRecords(
		fromCSV,
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		func(message string) {
			fmt.Println(message) //nolint:revive // has warning messages
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "complex test",
			args: args{
				filialID:       "OOO HealthCare",
				dominoSchedule: dominoSchedule,
				log: func(message string) {
					t.Error(message)
				},
			},
			want:    jsonUnmarshal(t, toJSON),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			schedule, err := service.CreateSchedule(tt.args.filialID, tt.args.dominoSchedule, tt.args.log)
			if err != nil {
				t.Errorf("CreateSchedule() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if schedule.IsEmpty() {
				t.Error("IsEmpty() == true")
			}

			got, err := schedule.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(jsonUnmarshal(t, got), tt.want) {
				t.Errorf("CreateSchedule() got = %s, want %s", got, toJSON)
			}
		})
	}
}
