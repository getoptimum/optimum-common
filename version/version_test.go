package version

import (
	"testing"
)

func TestGetLastCommitHash(t *testing.T) {
	hash := GetLastCommitHash()
	if len(hash) > 0 && len(hash) != 7 {
		t.Fatalf("commit hash length = %d, want 7", len(hash))
	}
}

func TestExtractCommitHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Long commit", "v0.0.0-abcdef123456", "abcdef1"},
		{"Short commit", "v0.0.0-abc123", "abc123"},
		{"Dirty commit", "v0.0.0-abcdef123456+dirty", "abcdef1"},
		{"Only commit", "abcdef123456", "abcdef1"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractCommitHash(tt.input); got != tt.want {
				t.Errorf("extractCommitHash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
