package imagehost

import (
	"fmt"
	"testing"
)

func TestPostIMG(t *testing.T) {
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
			gotFp, err := PostIMG(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostIMG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotFp)
			fmt.Printf("Link Direct: %s\n", gotFp.Direct)
			fmt.Printf("Thumb Direct: %s\n", gotFp.Thumb)
		})
	}
}
