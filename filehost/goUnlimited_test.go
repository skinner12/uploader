package filehost

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGoUnlimitedUpload(t *testing.T) {
	// Log level
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")

	type args struct {
		path   string
		apikey string
	}
	tests := []struct {
		name    string
		args    args
		wantS   LinkFileUploaded
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "File Upload",
			args: args{
				path:   "/home/tux/Downloads/amo_famo_na_pecorina.mp4",
				apikey: "30034xgh0ohqugf6kpqvd",
			},
			wantS: LinkFileUploaded{
				HTMLCode:    "htrr",
				EmbededCode: "asdassa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, err := GOUnlimited(tt.args.path, tt.args.apikey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS.EmbededCode != tt.wantS.EmbededCode {
				t.Errorf("Upload() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func TestExtractID(t *testing.T) {
	type args struct {
		element string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "IDExample",
			args: args{
				element: `<HTML><BODY><Form name='F1' action='https://gounlimited.to/' target='_parent' method='POST'><textarea name="op">upload_result</textarea><textarea name="fn">tjqusux44nbn</textarea><textarea name="st">OK</textarea></Form><Script>document.location='javascript:false';document.F1.submit();</Script></BODY></HTML>`,
			},
			want: "tjqusux44nbn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractID(tt.args.element); got != tt.want {
				t.Errorf("ExtractID() = %v, want %v", got, tt.want)
			}
		})
	}
}
