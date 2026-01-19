package generator

import (
	"testing"
)

func TestParseLengthsFromRegex(t *testing.T) {
	tests := []struct {
		regex   string
		wantMin int
		wantMax int
	}{
		{"[a-z]{6,10}", 6, 10},
		{"[A-Z]{8}", 8, 8},
		{"[0-9]{1,3}", 1, 3},
		{"[a-z]{5,12}", 5, 12},
		{"[a-z]", 1, 1},
	}

	for _, tt := range tests {
		min, max := ParseLengthsFromRegex(tt.regex)
		if min != tt.wantMin || max != tt.wantMax {
			t.Errorf("ParseLengthsFromRegex(%q) = (%d, %d), want (%d, %d)", tt.regex, min, max, tt.wantMin, tt.wantMax)
		}
	}
}

func TestParseCharsetFromRegex(t *testing.T) {
	tests := []struct {
		regex string
		want  string
	}{
		{"[a-c]", "abc"},
		{"[A-B0-1]", "AB01"},
		{"[#_]", "#_"},
		{"[a-zA-Z0-9!@#_-]", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#_-"},
	}

	for _, tt := range tests {
		got := ParseCharsetFromRegex(tt.regex)
		if got != tt.want {
			t.Errorf("ParseCharsetFromRegex(%q) = %q, want %q", tt.regex, got, tt.want)
		}
	}
}
