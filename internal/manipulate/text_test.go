package manipulate

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestDateInRange(t *testing.T) {

	// Log level
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")

	type args struct {
		dateFormat string
		dateString string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"older",
			args{
				time.RFC3339,
				"2020-02-28T22:53:20+02:00",
			},
			false,
			false,
		},
		{
			"in rnage",
			args{
				time.RFC3339,
				"2020-04-28T22:53:20+02:00",
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DateInRange(tt.args.dateFormat, tt.args.dateString)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateInRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DateInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateInRangeDays(t *testing.T) {
	// Log level
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")

	type args struct {
		dateCheck string
		days      int
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"older",
			args{
				"2020-02-28T22:53:20+02:00",
				10,
			},
			false,
			false,
		},
		{
			"in range",
			args{
				"2020-11-28T22:53:20+02:00",
				10,
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DateInRangeDays(tt.args.dateCheck, tt.args.days)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateInRangeDays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DateInRangeDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
