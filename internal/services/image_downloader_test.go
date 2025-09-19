package services

import (
	"testing"
)

// TestIsPrivateIP_DirectIPs verifies private address detection on direct IP inputs
func TestIsPrivateIP_DirectIPs(t *testing.T) {
	id := &ImageDownloader{}

	tests := []struct {
		host       string
		expectPriv bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		priv, err := id.isPrivateIP(tt.host)
		if err != nil {
			t.Fatalf("unexpected error for host %q: %v", tt.host, err)
		}
		if priv != tt.expectPriv {
			t.Errorf("isPrivateIP(%q) = %v, want %v", tt.host, priv, tt.expectPriv)
		}
	}
}

