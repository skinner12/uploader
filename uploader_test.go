package upload

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestUploadFile(t *testing.T) {

	username := os.Getenv("username")
	password := os.Getenv("password")
	fileToUpload := os.Getenv("fileToUpload")
	fileHost := os.Getenv("fileHost")
	uploaderPath := os.Getenv("fileHost")

	// Log level
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")
	type args struct {
		f        string
		path     string
		filehost string
		username string
		password string
		minSpeed string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"upload1",
			args{
				fileToUpload,
				uploaderPath,
				fileHost,
				username,
				password,
				"150k",
			},
			false,
		},
	}
	for _, tt := range tests {
		var l string
		var err error
		t.Run(tt.name, func(t *testing.T) {
			if username == "" || password == "" || fileToUpload == "" || uploaderPath == "" || fileHost == "" {
				t.Errorf("Username/Password/FilePath/UploaderPath/Filehost not in ENV. Add before run again this Test. Use 'export username=xxx.....'")
				t.FailNow()
			}
			if l, err = ToFileHost(tt.args.f, tt.args.path, tt.args.filehost, tt.args.username, tt.args.password, tt.args.minSpeed); (err != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			log.Infof("Link Uploaded: %s\n", l)
		})
	}
}
