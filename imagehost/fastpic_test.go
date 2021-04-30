package imagehost

import (
	"fmt"
	"testing"
)

func TestFastPic(t *testing.T) {
	type args struct {
		f string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			"Image1",
			args{
				"/home/tux/Downloads/64770_1080.jpg",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pic, err := FastPic(tt.args.f)
			if err != nil {
				t.Error(err)
			}
			fmt.Println(pic)
		})
	}
}
