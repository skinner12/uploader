package imagehost

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestPostImages(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")
	type args struct {
		f string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"file",
			args{
				"/home/tux/Downloads/20200510_100841.jpg",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFp, err := PostImages(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotFp)
			fmt.Printf("Link Direct: %s\n", gotFp.Direct)
			fmt.Printf("Thumb Direct: %s\n", gotFp.Thumb)
		})
	}
}
