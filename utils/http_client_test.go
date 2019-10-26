package utils

import (
	"reflect"
	"testing"
)

func TestPost(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *httpClient
	}{
		// TODO: Add test cases.
		struct {
			name string
			args args
			want *httpClient
		}{name: "hello", args: args{"http://baidu.com"}, want: NewHttpClient()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Post(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Post() = %v, want %v", got, tt.want)
			}
		})
	}
}
