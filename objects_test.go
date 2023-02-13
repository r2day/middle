package middle

import (
	"reflect"
	"testing"
)

func TestDumpLoginInfo(t *testing.T) {
	type args struct {
		namespace string
		user      string
		avatar    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test",
			args{
				"a",
				"b",
				"c",
			},
			"eyJuYW1lc3BhY2UiOiJhIiwidXNlciI6ImIiLCJhdmF0YXIiOiJjIn0=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DumpLoginInfo(tt.args.namespace, tt.args.user, tt.args.avatar); got != tt.want {
				t.Errorf("LoadLoginInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadLoginInfo(t *testing.T) {
	ob := &LoginInfo{
		"a",
		"b",
		"c",
	}
	type args struct {
		payload string
	}
	tests := []struct {
		name string
		args args
		want *LoginInfo
	}{
		// TODO: Add test cases.
		{
			"test",
			args{
				"eyJuYW1lc3BhY2UiOiJhIiwidXNlciI6ImIiLCJhdmF0YXIiOiJjIn0=",
			},
			ob,

		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadLoginInfo(tt.args.payload); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadLoginInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
