package qiosid

import (
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{1, "QIOS-000001"},
		{42, "QIOS-000042"},
		{999_999, "QIOS-999999"},
		{1_000_000, "QIOS-1000000"}, // overflow tetap render — schema VARCHAR(20) cukup
	}
	for _, c := range cases {
		got := Format(c.in)
		if got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseSequence(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"QIOS-000001", 1, false},
		{"QIOS-1", 1, false},
		{"QIOS-999999", 999_999, false},
		{"qios-0000000001", 0, true}, // wrong-case prefix
		{"QM-000001", 0, true},       // old format
		{"QIOS-", 0, true},
		{"QIOS-abc", 0, true},
		{"", 0, true},
	}
	for _, c := range cases {
		got, err := parseSequence(c.in)
		gotErr := err != nil
		if gotErr != c.wantErr {
			t.Errorf("parseSequence(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("parseSequence(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
