package dominocsv_test

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"reflect"
	"testing"

	"prodoctorov/internal/service/dominocsv"
)

func TestDecodeValue(t *testing.T) {
	type args struct {
		text string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sample 1",
			args: args{
				text: "&#1040;&#1083;&#1083;&#1077;&#1088;&#1075;&#1086;&#1083;&#1086;&#1075;",
			},
			want: "Аллерголог",
		},
		{
			name: "sample 2",
			args: args{
				text: "&#1048;&#1074;&#1072;&#1085;&#1086;&#1074; &#1048;.&#1048;.",
			},
			want: "Иванов И.И.",
		},
		{
			name: "sample 3",
			args: args{
				text: "4 &#1082;&#1072;&#1073;&#1080;&#1085;&#1077;&#1090;",
			},
			want: "4 кабинет",
		},
		{
			name: "sample 4",
			args: args{
				text: "31.7.21",
			},
			want: "31.7.21",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := dominocsv.DecodeValue(tt.args.text)
			if got != tt.want {
				t.Errorf("DecodeValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	//go:embed testdata.csv
	testCSV []byte
)

func TestReader_Read(t *testing.T) {
	reader, err := dominocsv.NewReader(bytes.NewReader(testCSV))
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}

	tests := []struct {
		name    string
		want    []string
		wantErr error
	}{
		{
			name:    "line 1",
			want:    []string{"spec", "name", "cell", "duration", "free", "room", ""},
			wantErr: nil,
		},
		{
			name:    "line 2",
			want:    []string{"Гастроэнтеролог", "Иванов Е.А.", "1.4.20 10:00:00", "30", "free", "", ""},
			wantErr: nil,
		},
		{
			name:    "line 3",
			want:    []string{"Дерматолог", "Иванов Е.С.", "1.7.21 10:00:00", "30", "busy", "6 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 4",
			want:    []string{"ЛОР", "Иванов Л.Б.", "1.10.19 10:00:00", "30", "free", "11 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 5",
			want:    []string{"Массажист", "Иванов И.Н.", "5.8.19 10:00:00", "30", "free", "", ""},
			wantErr: nil,
		},
		{
			name:    "line 6",
			want:    []string{"Невролог", "Иванов Д.М.", "12.7.21 10:00:00", "30", "busy", "13 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 7",
			want:    []string{"Терапевт", "Петров Е.Е.", "20.6.17 10:00:00", "30", "free", "", ""},
			wantErr: nil,
		},
		{
			name:    "line 8",
			want:    []string{"УЗИ", "Петров Г.Б.", "24.7.21 10:00:00", "", "free", "", ""},
			wantErr: nil,
		},
		{
			name:    "line 9",
			want:    []string{"Уролог", "Петров Г.А.", "1.7.21 10:00:00", "", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 10",
			want:    []string{"Уролог", "Петров Г.А.", "1.7.21 10:20:00", "", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 11",
			want:    []string{"Уролог", "Петров Г.А.", "1.7.21 10:40:00", "", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 12",
			want:    []string{"Уролог", "Петров Г.А.", "1.7.21 11:00:00", "", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 13",
			want:    []string{"Хирург", "Петров Д.А.", "1.11.21 9:00:00", "30", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 14",
			want:    []string{"Хирург", "Петров Д.А.", "1.11.21 9:30:00", "30", "busy", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 15",
			want:    []string{"Хирург", "Петров Д.А.", "1.11.21 10:00:00", "30", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "line 16",
			want:    []string{"Хирург", "Петров Д.А.", "1.11.21 10:30:00", "30", "free", "12 кабинет", ""},
			wantErr: nil,
		},
		{
			name:    "end",
			want:    nil,
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := reader.Read()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Read() got error = %v, wantErr = %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
