package filehost

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestK2S(t *testing.T) {
	// Log level
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")
	type args struct {
		path     string
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantL   string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "k2s",
			args: args{
				path:     "/home/tux/Downloads/VGGG3.png",
				username: "mazzu83@gmail.com",
				password: "Marlbor0!",
			},
			wantL:   "/home/tux/Downloads/VGGG3.png",
			wantErr: false,
		},
		/*{
			name: "k2s",
			args: args{
				path:     "/home/tux/Downloads/dummyfile",
				username: "mazzu83@gmail.com",
				password: "Marlbor0!",
			},
			wantL:   "/home/tux/Downloads/VGGG3.png",
			wantErr: false,
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotL, err := K2S(tt.args.path, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("K2S() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(gotL, "https://k2s.cc/file") {
				t.Errorf("K2S() = %v, want %v", gotL, tt.wantL)
			}
		})
	}
}
