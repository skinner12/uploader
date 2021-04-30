package manipulate

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestChangeMD5(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "md5",
			args: args{
				path: "/home/tux/Downloads/Lato2OFL.zip",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChangeMD5(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("ChangeMD5() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveSpaceFromFile(t *testing.T) {
	type args struct {
		folder string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				folder: "/home/tux/Downloads/aaa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RemoveSpaceFromFile(tt.args.folder); (err != nil) != tt.wantErr {
				t.Errorf("RemoveSpaceFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//
func TestRemoveSpaceFromDir(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.Infoln("DebugLevel LOG Set")

	dir := "/home/tux/Downloads/aaa/"
	// Rename files, delete space and symbols
	if err := RemoveSpaceFromFile(dir); err != nil {
		log.WithFields(log.Fields{
			"Err": err,
		}).Errorln("TorrentLAB[Upload] - Rename files")
		t.Fatal(err)
	}

	log.Infof("Renamed all files successfull for %s", dir)

}
