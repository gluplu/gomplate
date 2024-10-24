package azure

import (
	"testing"
)

func Test_extractTagValue(t *testing.T) {
	type args struct {
		tag    string
		result string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "some tag values", args: args{tag: "blah", result: "blah:somevalue"}, want: "somevalue"},
		{name: "multiple tag values", args: args{tag: "blah", result: "blah:somevalue;kukta:someothervalue"}, want: "somevalue"},
		{name: "some unrelated tags", args: args{tag: "blah", result: "nonblah:somevalue;kukta:someothervalue"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTagValue(tt.args.tag, tt.args.result); got != tt.want {
				t.Errorf("extractTagValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractJsonValue(t *testing.T) {
	type args struct {
		json string
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "external IP", args: args{json: `{"loadbalancer":{"publicIpAddresses":[{"frontendIpAddress":"20.55.51.220","privateIpAddress":"10.1.0.4"}],"inboundRules":[],"outboundRules":[]}}`, path: "loadbalancer/publicIpAddresses/0/frontendIpAddress"}, want: "20.55.51.220"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractJsonValue(tt.args.json, tt.args.path); got != tt.want {
				t.Errorf("extractJsonValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
