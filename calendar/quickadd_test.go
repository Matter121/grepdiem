package calendar

import (
	"reflect"
	"testing"
	"time"
)

func TestParseEvent(t *testing.T) {
	selected := d("2026-10-07") // a Wednesday
	tueThu := []time.Weekday{time.Tuesday, time.Thursday}
	cases := []struct {
		line string
		want EventSpec
	}{
		{"Dentist", EventSpec{Title: "Dentist"}},
		{"Dentist 10:30", EventSpec{Title: "Dentist", Start: "10:30"}},
		{"Lunch with Sam 12", EventSpec{Title: "Lunch with Sam", Start: "12:00"}},
		{"Lecture 9-11", EventSpec{Title: "Lecture", Start: "09:00", End: "11:00"}},
		{"Lecture 8:15-10", EventSpec{Title: "Lecture", Start: "08:15", End: "10:00"}},
		{"Study tue thu", EventSpec{Title: "Study", Weekdays: tueThu}},
		{"Study Thursday TUE", EventSpec{Title: "Study", Weekdays: tueThu}},
		{"Study tue,thu", EventSpec{Title: "Study", Weekdays: tueThu}},
		// A range of weekdays runs forwards through the week.
		{"Work mon-sat 9-17", EventSpec{Title: "Work", Start: "09:00", End: "17:00", Weekdays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}}},
		{"Trip Fri-Mon", EventSpec{Title: "Trip", Weekdays: []time.Weekday{
			time.Monday, time.Friday, time.Saturday, time.Sunday}}},
		{"Gym mon-wed,fri", EventSpec{Title: "Gym", Weekdays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Friday}}},
		{"Shift mon-", EventSpec{Title: "Shift mon-"}},
		{"Study tue thu 13-15 until w50",
			EventSpec{Title: "Study", Start: "13:00", End: "15:00", Weekdays: tueThu, Until: d("2026-12-13")}},
		{"Study 13-15 tue thu until 2026-12-13",
			EventSpec{Title: "Study", Start: "13:00", End: "15:00", Weekdays: tueThu, Until: d("2026-12-13")}},
		{"Study tue thu until 13 dec", EventSpec{Title: "Study", Weekdays: tueThu, Until: d("2026-12-13")}},
		// An end that has already passed this year means next year.
		{"Study tue until w3", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-01-24")}},
		{"Study tue until week 1", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-01-10")}},
		{"Study tue until Week 50", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue until 261213", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue until 13/12", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue until 13.12", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue until 5/2", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-02-05")}},
		{"Study tue until 13/12/27", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-12-13")}},
		{"Study tue until 13.12.2027", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-12-13")}},
		{"Study tue until dec 13", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue until 13 dec 2027", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-12-13")}},
		{"Study tue until 31/2", EventSpec{Title: "Study tue until 31/2"}},
		{"Study tue until 13/12/127", EventSpec{Title: "Study tue until 13/12/127"}},
		// "u" is short for "until".
		{"Study tue u w50", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")}},
		{"Study tue thu 13-15 U 13 dec",
			EventSpec{Title: "Study", Start: "13:00", End: "15:00", Weekdays: tueThu, Until: d("2026-12-13")}},
		{"Thank u", EventSpec{Title: "Thank u"}},
		{"Study tue until 261332", EventSpec{Title: "Study tue until 261332"}},
		{"Study tue until 5 feb", EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2027-02-05")}},
		// An end without weekdays repeats on the selected day's weekday.
		{"Swim until w45", EventSpec{Title: "Swim", Weekdays: []time.Weekday{time.Wednesday}, Until: d("2026-11-08")}},
		// Keywords only count at the end of the line.
		{"Sun cream 13", EventSpec{Title: "Sun cream", Start: "13:00"}},
		{"Buy 13 eggs", EventSpec{Title: "Buy 13 eggs"}},
		{"Read until dawn", EventSpec{Title: "Read until dawn"}},
		// The first word is always title.
		{"Sun 13", EventSpec{Title: "Sun", Start: "13:00"}},
		{"13", EventSpec{Title: "13"}},
		// Quotes force a literal title.
		{`"Apollo 13" fri 20`, EventSpec{Title: "Apollo 13", Start: "20:00", Weekdays: []time.Weekday{time.Friday}}},
		{`"Apollo 13"`, EventSpec{Title: "Apollo 13"}},
		// Not times.
		{"Room 25", EventSpec{Title: "Room 25"}},
		{"Call 12:75", EventSpec{Title: "Call 12:75"}},
		{"  Dentist   10  ", EventSpec{Title: "Dentist", Start: "10:00"}},
		{"", EventSpec{}},
	}
	for _, tc := range cases {
		got := ParseEvent(tc.line, selected)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ParseEvent(%q)\n got %+v\nwant %+v", tc.line, got, tc.want)
		}
	}
}
