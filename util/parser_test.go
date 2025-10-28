package util

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "Test seconds",
			input:   "10s",
			want:    10 * time.Second,
			wantErr: false,
		},
		{
			name:    "Test minutes",
			input:   "2m",
			want:    2 * time.Minute,
			wantErr: false,
		},
		{
			name:    "Test hours",
			input:   "3h",
			want:    3 * time.Hour,
			wantErr: false,
		},
		{
			name:    "Test days",
			input:   "4d",
			want:    4 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "Test weeks",
			input:   "5w",
			want:    5 * 7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "Test invalid format",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
