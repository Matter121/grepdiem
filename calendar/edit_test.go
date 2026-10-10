package calendar

import (
	"testing"
	"time"
)

var today = d("2026-10-09")

func TestEditingARecurringEventChangesEveryOccurrence(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	c.UpdateEvent(e.ID, EventSpec{Title: "Study group", Start: "13:00", Weekdays: []time.Weekday{time.Tuesday}})

	for _, date := range []string{"2026-10-06", "2026-11-03"} {
		got := c.Day(d(date), today).Rows
		if len(got) != 1 || got[0].Title() != "Study group" || got[0].Time() != "13:00" {
			t.Errorf("%s: %+v", date, got)
		}
	}
}

func TestEditKeepsSkipsAndColor(t *testing.T) {
	c := New()
	c.AddEvent(today, EventSpec{Title: "Other"})
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	c.Skip(e.ID, d("2026-10-13"))
	c.UpdateEvent(e.ID, EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday, time.Thursday}})

	if n := len(c.Day(d("2026-10-13"), today).Rows); n != 0 {
		t.Errorf("skipped Tuesday came back")
	}
	got := c.Day(d("2026-10-15"), today).Rows
	if len(got) != 1 || got[0].Event.Color != e.Color {
		t.Errorf("new Thursday: %+v, want colour %d", got, e.Color)
	}
}

func TestRemovingWeekdaysCollapsesToOneEventOnTheStartDate(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-12-13")})
	c.UpdateEvent(e.ID, EventSpec{Title: "Study"})

	if n := len(c.Day(d("2026-10-06"), today).Rows); n != 1 {
		t.Errorf("start date has %d events, want 1", n)
	}
	if n := len(c.Day(d("2026-10-13"), today).Rows); n != 0 {
		t.Errorf("later Tuesday still has the event")
	}
}

func TestAddingWeekdaysMakesAOneOffRecurFromItsDate(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-07"), EventSpec{Title: "Study"})
	c.UpdateEvent(e.ID, EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Thursday}})

	if n := len(c.Day(d("2026-10-01"), today).Rows); n != 0 {
		t.Errorf("Thursday before its date has the event")
	}
	if n := len(c.Day(d("2026-10-08"), today).Rows); n != 1 {
		t.Errorf("first Thursday after its date is missing")
	}
}

func TestMoveOneOffEvent(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-14"), EventSpec{Title: "Dentist"})
	if !c.MoveEvent(e.ID, d("2026-10-21")) {
		t.Fatal("move refused")
	}
	if n := len(c.Day(d("2026-10-14"), today).Rows); n != 0 {
		t.Errorf("still on the old day")
	}
	if n := len(c.Day(d("2026-10-21"), today).Rows); n != 1 {
		t.Errorf("not on the new day")
	}
}

func TestRecurringEventCannotBeMoved(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	if c.MoveEvent(e.ID, d("2026-10-07")) {
		t.Error("move of a Recurring Event was accepted")
	}
	if n := len(c.Day(d("2026-10-06"), today).Rows); n != 1 {
		t.Errorf("event changed despite refusal")
	}
}

func TestDeleteEventRemovesEveryOccurrence(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	c.DeleteEvent(e.ID)
	if n := len(c.Day(d("2026-10-13"), today).Rows); n != 0 {
		t.Errorf("still there")
	}
}

func TestNewEventsGetDifferentColorsAndColorCycles(t *testing.T) {
	c := New()
	a := c.AddEvent(today, EventSpec{Title: "A"})
	b := c.AddEvent(today, EventSpec{Title: "B"})
	if a.Color == b.Color {
		t.Errorf("both got colour %d", a.Color)
	}
	for range Colors {
		c.CycleColor(a.ID)
	}
	if got := c.Day(today, today).Rows[0].Event.Color; got != a.Color {
		t.Errorf("after a full cycle colour = %d, want %d", got, a.Color)
	}
	c.CycleColor(a.ID)
	if got := c.Day(today, today).Rows[0].Event.Color; got == a.Color {
		t.Errorf("colour did not change")
	}
}

func TestFirstOccurrence(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-07"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday, time.Thursday}})
	if got := e.First(); got != d("2026-10-08") {
		t.Errorf("first = %v, want 2026-10-08", got)
	}
	never := c.AddEvent(d("2026-10-07"), EventSpec{Title: "X", Weekdays: []time.Weekday{time.Tuesday}, Until: d("2026-10-08")})
	if got := never.First(); !got.IsZero() {
		t.Errorf("event that ends before its first weekday: first = %v", got)
	}
}

func TestCloneIsIndependent(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	task := c.AddTask(today, "Buy milk")
	snapshot := c.Clone()

	c.Skip(e.ID, d("2026-10-13"))
	c.DeleteTask(task.ID)
	c.AddEvent(today, EventSpec{Title: "New"})

	if n := len(snapshot.Day(d("2026-10-13"), today).Rows); n != 1 {
		t.Errorf("snapshot lost the skipped Occurrence")
	}
	day := snapshot.Day(today, today)
	if len(day.Rows) != 1 || day.Rows[0].Kind != TaskRow {
		t.Errorf("snapshot changed: %+v", day)
	}
}
