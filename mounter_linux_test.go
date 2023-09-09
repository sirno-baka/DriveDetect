package DriveDetect

import (
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "usb drive connected",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectAndMount()
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("Got", got)
		})
	}
}
