package calendar

import (
	"reflect"
	"testing"
	"time"
)

func TestDescribe(t *testing.T) {
	selected := d("2026-10-07")
	cases := []struct {
		spec EventSpec
		want string
	}{
		{EventSpec{Title: "Dentist"}, "Dentist · Wed 7 Oct"},
		{EventSpec{Title: "Dentist", Start: "10:30"}, "Dentist · Wed 7 Oct · 10:30"},
		{EventSpec{Title: "Study", Start: "13:00", End: "15:00",
			Weekdays: []time.Weekday{time.Tuesday, time.Thursday}, Until: d("2026-12-13")},
			"Study · Tue, Thu · 13:00–15:00 · until Sun 13 Dec"},
		{EventSpec{Title: "Gym", Weekdays: []time.Weekday{time.Monday}}, "Gym · Mon · no end"},
		{EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-01-24")},
			"Study · Tue · until Sun 24 Jan 2027"},
	}
	for _, tc := range cases {
		if got := Describe(tc.spec, selected); got != tc.want {
			t.Errorf("got  %q\nwant %q", got, tc.want)
		}
	}
}

func TestLineRecreatesTheEvent(t *testing.T) {
	selected := d("2026-10-07")
	cases := []struct {
		spec EventSpec
		want string
	}{
		{EventSpec{Title: "Dentist", Start: "10:30"}, "Dentist 10:30"},
		{EventSpec{Title: "Study", Start: "13:00", End: "15:00",
			Weekdays: []time.Weekday{time.Tuesday, time.Thursday}, Until: d("2026-12-13")},
			"Study tue thu 13-15 until 2026-12-13"},
		{EventSpec{Title: "Apollo 13"}, `"Apollo 13"`},
		{EventSpec{Title: "Sun", Start: "09:00"}, "Sun 9"},
		{EventSpec{Title: "Free fri", Weekdays: []time.Weekday{time.Friday}}, `"Free fri" fri`},
	}
	for _, tc := range cases {
		line := Line(tc.spec)
		if line != tc.want {
			t.Errorf("Line = %q, want %q", line, tc.want)
		}
		if back := ParseEvent(line, selected); !reflect.DeepEqual(back, tc.spec) {
			t.Errorf("%q parses back to %+v, want %+v", line, back, tc.spec)
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
