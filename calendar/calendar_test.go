package calendar

import (
	"testing"
	"time"
)

func d(s string) Date {
	date, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return date
}

func titles(day Day) []string {
	var out []string
	for _, r := range day.Rows {
		if r.Kind == EventRow {
			out = append(out, r.Title())
		}
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestOneOffEventShowsOnlyOnItsDate(t *testing.T) {
	c := New()
	c.AddEvent(d("2026-10-14"), EventSpec{Title: "Dentist"})

	today := d("2026-10-09")
	if got := titles(c.Day(d("2026-10-14"), today)); !equal(got, []string{"Dentist"}) {
		t.Errorf("on its date: got %v", got)
	}
	if got := titles(c.Day(d("2026-10-15"), today)); len(got) != 0 {
		t.Errorf("day after: got %v", got)
	}
}

func TestRecurringEventOccurrences(t *testing.T) {
	c := New()
	// Added on Wed 7 Oct 2026, Tue+Thu, until Sun 13 Dec 2026.
	e := c.AddEvent(d("2026-10-07"), EventSpec{
		Title:    "Study",
		Weekdays: []time.Weekday{time.Tuesday, time.Thursday},
		Until:    d("2026-12-13"),
	})
	c.Skip(e.ID, d("2026-10-15"))

	today := d("2026-10-09")
	cases := []struct {
		date string
		want bool
		why  string
	}{
		{"2026-10-06", false, "Tuesday before the start date"},
		{"2026-10-07", false, "start date is a Wednesday"},
		{"2026-10-08", true, "first Thursday on or after start"},
		{"2026-10-13", true, "a Tuesday"},
		{"2026-10-14", false, "a Wednesday"},
		{"2026-10-15", false, "skipped Thursday"},
		{"2026-10-20", true, "Tuesday after the skip"},
		{"2026-12-10", true, "last Thursday before the end"},
		{"2026-12-15", false, "Tuesday after the end"},
	}
	for _, tc := range cases {
		got := len(c.Day(d(tc.date), today).Rows) == 1
		if got != tc.want {
			t.Errorf("%s (%s): occurs = %v, want %v", tc.date, tc.why, got, tc.want)
		}
	}
}

func TestRecurringEventWithoutEndGoesOn(t *testing.T) {
	c := New()
	c.AddEvent(d("2026-10-07"), EventSpec{Title: "Gym", Weekdays: []time.Weekday{time.Monday}})
	if got := titles(c.Day(d("2031-03-03"), d("2026-10-09"))); !equal(got, []string{"Gym"}) {
		t.Errorf("Monday years later: got %v", got)
	}
}

func TestEventsSortAllDayFirstThenByTime(t *testing.T) {
	c := New()
	day := d("2026-10-09")
	c.AddEvent(day, EventSpec{Title: "Lunch", Start: "12:00"})
	c.AddEvent(day, EventSpec{Title: "Birthday"})
	c.AddEvent(day, EventSpec{Title: "Standup", Start: "09:15"})

	want := []string{"Birthday", "Standup", "Lunch"}
	if got := titles(c.Day(day, day)); !equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDateWeekdayAndISOWeek(t *testing.T) {
	date := d("2026-12-13")
	if date.Weekday() != time.Sunday {
		t.Errorf("weekday = %v", date.Weekday())
	}
	if _, w := date.ISOWeek(); w != 50 {
		t.Errorf("week = %d, want 50", w)
	}
	if got := d("2026-01-31").AddMonths(1); got != d("2026-02-28") {
		t.Errorf("31 Jan + 1 month = %v, want 2026-02-28", got)
	}
}
