package domino_test

import (
	"reflect"
	"testing"
	"time"

	"prodoctorov/internal/service/domino"
)

func TestNewDominoRecord(t *testing.T) {
	type args struct {
		record  []string
		timeNow time.Time
	}

	tests := []struct {
		name    string
		args    args
		want    *domino.Record
		wantErr bool
	}{
		{
			name: "record 1",
			args: args{
				record:  []string{"Гастроэнтеролог", "Иванов Е.А.", "1.4.20 10:00:00", "30", "free", "", ""},
				timeNow: time.Date(2020, 04, 10, 0, 0, 0, 0, time.UTC),
			},
			want: &domino.Record{
				Spec:      "Гастроэнтеролог",
				Name:      "Иванов Е.А.",
				StartTime: time.Date(2020, 04, 01, 10, 0, 0, 0, time.UTC),
				Free:      true,
				Duration:  30 * time.Minute,
				Room:      "",
			},
			wantErr: false,
		},
		{
			name: "record 2",
			args: args{
				record:  []string{"Дерматолог", "Иванов Е.С.", "1.7.21 10:00:00", "20", "busy", "6 кабинет", ""},
				timeNow: time.Date(2021, 07, 05, 0, 0, 0, 0, time.UTC),
			},
			want: &domino.Record{
				Spec:      "Дерматолог",
				Name:      "Иванов Е.С.",
				StartTime: time.Date(2021, 07, 01, 10, 0, 0, 0, time.UTC),
				Free:      false,
				Duration:  20 * time.Minute,
				Room:      "6 кабинет",
			},
			wantErr: false,
		},
		{
			name: "empty duration",
			args: args{
				record:  []string{"Дерматолог", "Иванов Е.С.", "1.7.21 10:30:00", "", "busy", "6 кабинет", ""},
				timeNow: time.Date(2021, 07, 05, 0, 0, 0, 0, time.UTC),
			},
			want: &domino.Record{
				Spec:      "Дерматолог",
				Name:      "Иванов Е.С.",
				StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
				Free:      false,
				Duration:  0,
				Room:      "6 кабинет",
			},
			wantErr: false,
		},
		{
			name: "record expired",
			args: args{
				record:  []string{"Гастроэнтеролог", "Иванов Е.А.", "1.4.20 10:00:00", "30", "free", "", ""},
				timeNow: time.Date(2021, 07, 05, 0, 0, 0, 0, time.UTC),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := domino.NewRecord(tt.args.record, tt.args.timeNow)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRecord() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRecord() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecords_Cleaned(t *testing.T) {
	tests := []struct {
		name    string
		records domino.Records
		want    domino.Records
	}{
		{
			name:    "empty",
			records: make(domino.Records, 0),
			want:    make(domino.Records, 0),
		},
		{
			name: "one record w/o duration",
			records: domino.Records{&domino.Record{
				Duration: 0,
			}},
			want: domino.Records{&domino.Record{
				Duration: domino.DefaultMeetDuration,
			}},
		},
		{
			name: "two records w/ duration",
			records: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
			want: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
		},
		{
			name: "two records w/o duration",
			records: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  0,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  0,
				},
			},
			want: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
		},
		{
			name: "last record w/o duration",
			records: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 11, 00, 0, 0, time.UTC),
					Duration:  0,
				},
			},
			want: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 11, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
		},
		{
			name: "last record w/o duration && next day",
			records: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 02, 10, 00, 0, 0, time.UTC),
					Duration:  0,
				},
			},
			want: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 30, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 02, 10, 00, 0, 0, time.UTC),
					Duration:  20 * time.Minute,
				},
			},
		},
		{
			name: "meetings rarely",
			records: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  0,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 05, 10, 00, 0, 0, time.UTC),
					Duration:  0,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 10, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
			want: domino.Records{
				&domino.Record{
					StartTime: time.Date(2021, 07, 01, 10, 00, 0, 0, time.UTC),
					Duration:  20 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 05, 10, 00, 0, 0, time.UTC),
					Duration:  20 * time.Minute,
				},
				&domino.Record{
					StartTime: time.Date(2021, 07, 10, 10, 00, 0, 0, time.UTC),
					Duration:  30 * time.Minute,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := tt.records.Cleaned(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cleaned() = %v, want %v", got, tt.want)

				for i, r := range got {
					t.Errorf("Got item %d: %v", i, r)
				}

				for i, r := range tt.want {
					t.Errorf("Want item %d: %v", i, r)
				}
			}
		})
	}
}
