package output

import (
	"testing"
)

func TestConvertXeroDate_Standard(t *testing.T) {
	got := convertXeroDate("/Date(1639094400000+0000)/")
	want := "2021-12-10"
	if got != want {
		t.Errorf("convertXeroDate(/Date(1639094400000+0000)/) = %q, want %q", got, want)
	}
}

func TestConvertXeroDate_WithoutOffset(t *testing.T) {
	// Some Xero dates omit the timezone offset
	got := convertXeroDate("/Date(1639094400000)/")
	want := "2021-12-10"
	if got != want {
		t.Errorf("convertXeroDate(/Date(1639094400000)/) = %q, want %q", got, want)
	}
}

func TestConvertXeroDate_ZeroEpoch(t *testing.T) {
	got := convertXeroDate("/Date(0)/")
	want := "1970-01-01"
	if got != want {
		t.Errorf("convertXeroDate(/Date(0)/) = %q, want %q", got, want)
	}
}

func TestConvertXeroDate_NonXeroDate(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"2021-12-10"},
		{"not a date"},
		{""},
		{"2023-01-15T10:30:00Z"},
	}

	for _, tt := range tests {
		got := convertXeroDate(tt.input)
		if got != tt.input {
			t.Errorf("convertXeroDate(%q) = %q, want unchanged %q", tt.input, got, tt.input)
		}
	}
}

func TestConvertXeroDate_NegativeOffset(t *testing.T) {
	got := convertXeroDate("/Date(1639094400000-0500)/")
	// The offset in Xero dates is informational; the ms are always UTC
	want := "2021-12-10"
	if got != want {
		t.Errorf("convertXeroDate(/Date(1639094400000-0500)/) = %q, want %q", got, want)
	}
}
