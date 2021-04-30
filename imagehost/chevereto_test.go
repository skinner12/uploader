package imagehost

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestChevereto(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")
	type args struct {
		url    string
		apiKey string
		f      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"File Local",
			args{
				"https://image.amateurshouse.com",
				"d9a00319298c0eea7c00639922a985dff3f1f79a",
				"/home/tux/Downloads/20200510_100841.jpg",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFp, err := Chevereto(tt.args.url, tt.args.apiKey, tt.args.f)

			if (err != nil) != tt.wantErr {
				t.Errorf("Chevereto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotFp)
			fmt.Printf("Link Direct: %s\n", gotFp.Direct)
		})
	}
}
