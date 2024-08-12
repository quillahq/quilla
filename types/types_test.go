package types

import (
	"reflect"
	"testing"
	"time"
)

func TestVersion_String(t *testing.T) {
	type fields struct {
		Major      int64
		Minor      int64
		Patch      int64
		PreRelease string
		Metadata   string
		Original   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "semver with v",
			fields: fields{
				Major:    1,
				Minor:    1,
				Patch:    0,
				Original: "v1.1.0",
			},
			want: "v1.1.0",
		},
		{
			name: "semver standard",
			fields: fields{
				Major:    1,
				Minor:    1,
				Patch:    5,
				Original: "1.1.5",
			},
			want: "1.1.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.fields.Major,
				Minor:      tt.fields.Minor,
				Patch:      tt.fields.Patch,
				PreRelease: tt.fields.PreRelease,
				Metadata:   tt.fields.Metadata,
				Original:   tt.fields.Original,
			}
			if got := v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpired(t *testing.T) {
	aprv := Approval{
		Deadline: time.Now().Add(-5 * time.Second),
	}

	if !aprv.Expired() {
		t.Errorf("expected approval to be expired")
	}
}

func TestNotExpired(t *testing.T) {
	aprv := Approval{
		Deadline: time.Now().Add(5 * time.Second),
	}

	if aprv.Expired() {
		t.Errorf("expected approval to be not expired")

	}
}

func TestParseEventNotificationChannels(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no chans",
			args: args{map[string]string{"foo": "bar"}},
			want: []string{},
		},
		{
			name: "one chan",
			args: args{map[string]string{QuillaNotificationChanAnnotation: "verychan"}},
			want: []string{"verychan"},
		},
		{
			name: "two chans with space",
			args: args{map[string]string{QuillaNotificationChanAnnotation: "verychan, corp"}},
			want: []string{"verychan", "corp"},
		},
		{
			name: "two chans no space",
			args: args{map[string]string{QuillaNotificationChanAnnotation: "verychan,corp"}},
			want: []string{"verychan", "corp"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseEventNotificationChannels(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEventNotificationChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReleaseNotesURL(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil map",
			args: args{},
			want: "",
		},
		{
			name: "empty map",
			args: args{
				make(map[string]string),
			},
			want: "",
		},
		{
			name: "link",
			args: args{
				annotations: map[string]string{
					QuillaReleaseNotesURL: "http://link",
				},
			},
			want: "http://link",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseReleaseNotesURL(tt.args.annotations); got != tt.want {
				t.Errorf("ParseReleaseNotesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
