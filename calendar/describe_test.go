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
