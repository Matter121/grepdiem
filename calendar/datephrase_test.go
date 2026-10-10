package calendar

import (
	"strings"
	"testing"
)

func TestParseDatePhrase(t *testing.T) {
	selected := d("2026-10-07")
	cases := []struct {
		phrase string
		start  string // "" = not understood
		end    string // read as an end
	}{
		{"w50", "2026-12-07", "2026-12-13"},
		{"week 50", "2026-12-07", "2026-12-13"},
		// As an end, what is already past means the next one.
		{"w10", "2026-03-02", "2027-03-14"},
		{"3/2", "2026-02-03", "2027-02-03"},
		{"3 feb", "2026-02-03", "2027-02-03"},
		{"13.12", "2026-12-13", "2026-12-13"},
		{"dec 13", "2026-12-13", "2026-12-13"},
		// A date with its year is never moved.
		{"2026-03-01", "2026-03-01", "2026-03-01"},
		{"260301", "2026-03-01", "2026-03-01"},
		{"3.2.26", "2026-02-03", "2026-02-03"},
		{"3 feb 2026", "2026-02-03", "2026-02-03"},
		{"31 feb", "", ""},
		// A month alone is the first of the next such month.
		{"dec", "2026-12-01", "2026-12-01"},
		{"February", "2027-02-01", "2027-02-01"},
		{"oct", "2026-10-01", "2027-10-01"},
		{"2028", "", ""},
		{"3 feb 2026 please", "", ""},
		{"", "", ""},
	}
	for _, tc := range cases {
		for _, asEnd := range []bool{false, true} {
			want := tc.start
			if asEnd {
				want = tc.end
			}
			got, ok := parseDatePhrase(strings.Fields(tc.phrase), selected, asEnd)
			if want == "" {
				if ok {
					t.Errorf("parseDatePhrase(%q, end=%v) = %v, want not understood", tc.phrase, asEnd, got)
				}
				continue
			}
			if !ok || got != d(want) {
				t.Errorf("parseDatePhrase(%q, end=%v) = %v %v, want %s", tc.phrase, asEnd, got, ok, want)
			}
		}
	}
}

func TestParseGoto(t *testing.T) {
	selected := d("2026-10-07")
	cases := []struct {
		line string
		want string // "" = not understood
	}{
		{"14 oct", "2026-10-14"},
		{"3 feb", "2026-02-03"},
		// A month alone is the next such month; a year alone keeps the day.
		{"dec", "2026-12-01"},
		{"Oct", "2026-10-01"},
		{"february", "2027-02-01"},
		{"2028", "2028-10-07"},
		{"2025", "2025-10-07"},
		{"20280", ""},
		{"feb 3", "2026-02-03"},
		{"3 feb 2027", "2027-02-03"},
		{"3/2", "2026-02-03"},
		{"3.2.27", "2027-02-03"},
		{"14/10/2027", "2027-10-14"},
		{"31/2", ""},
		{"2027-03-01", "2027-03-01"},
		{"260101", "2026-01-01"},
		{"261022", "2026-10-22"},
		{"260230", ""},
		{"26-1022", ""},
		{"w50", "2026-12-07"},
		{" W1 ", "2025-12-29"},
		{"week 1", "2025-12-29"},
		{"week 60", ""},
		{"31 feb", ""},
		{"w60", ""},
		{"someday", ""},
		{"", ""},
	}
	for _, tc := range cases {
		got, ok := ParseGoto(tc.line, selected)
		if tc.want == "" {
			if ok {
				t.Errorf("ParseGoto(%q) = %v, want not understood", tc.line, got)
			}
			continue
		}
		if !ok || got != d(tc.want) {
			t.Errorf("ParseGoto(%q) = %v %v, want %s", tc.line, got, ok, tc.want)
		}
	}
}
