package calendar

import (
	"reflect"
	"testing"
	"time"
)

// row finds the Row with the given title on day, as listed on 2026-10-09.
func row(t *testing.T, c *Calendar, day Date, title string) Row {
	t.Helper()
	for _, r := range c.Day(day, today).Rows {
		if r.Title() == title {
			return r
		}
	}
	t.Fatalf("no Row %q on %v", title, day)
	return Row{}
}

func TestTickActsOnTheTaskOrSubtaskARowStandsFor(t *testing.T) {
	c := New()
	c.AddEvent(today, EventSpec{Title: "Dentist"})
	c.AddTask(today, "Pack", "passport", "charger")
	c.AddTask(today, "Buy milk")

	before := c.Clone()
	c.Tick(row(t, c, today, "Dentist"), today)
	if !reflect.DeepEqual(c, before) {
		t.Error("ticking an Event changed the Calendar")
	}

	c.Tick(row(t, c, today, "charger"), today)
	c.Tick(row(t, c, today, "Buy milk"), today)
	want := []line{
		{TaskRow, Open, "[ ] Pack 1/2"},
		{SubtaskRow, Open, "[ ] passport"},
		{SubtaskRow, Done, "[x] charger"},
		{TaskRow, Done, "[x] Buy milk"},
	}
	if got := lines(c.Day(today, today))[1:]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	c.Tick(row(t, c, today, "Buy milk"), today)
	if got := row(t, c, today, "Buy milk").State; got != Open {
		t.Errorf("ticking a Done Task again left it %v", got)
	}
}

func TestOnlyAnOccurrenceCanBeSkippedAndTheRestRemain(t *testing.T) {
	c := New()
	c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Tuesday}})
	c.AddEvent(d("2026-10-13"), EventSpec{Title: "Dentist"})
	c.AddTask(today, "Buy milk")

	study, dentist := row(t, c, d("2026-10-13"), "Study"), row(t, c, d("2026-10-13"), "Dentist")
	if !study.Occurrence() || dentist.Occurrence() || row(t, c, today, "Buy milk").Occurrence() {
		t.Error("only the Recurring Event's Row is an Occurrence")
	}

	c.Skip(dentist)
	c.Skip(study)
	if got := c.Day(d("2026-10-13"), today).Rows; len(got) != 1 || got[0].Title() != "Dentist" {
		t.Errorf("the skipped day lists %+v, want only Dentist", got)
	}
	if n := len(c.Day(d("2026-10-20"), today).Rows); n != 1 {
		t.Errorf("the next Occurrence is gone too")
	}
}

func TestDeleteTakesASubtaskAloneAndATaskWithItsSubtasks(t *testing.T) {
	c := New()
	c.AddTask(today, "Pack", "passport", "charger")
	c.AddTask(today, "Buy milk")

	c.Delete(row(t, c, today, "passport"))
	want := []line{{TaskRow, Open, "[ ] Pack 0/1"}, {SubtaskRow, Open, "[ ] charger"}, {TaskRow, Open, "[ ] Buy milk"}}
	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, want) {
		t.Errorf("after deleting a Subtask: %+v", got)
	}

	c.Delete(row(t, c, today, "Pack"))
	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, want[2:]) {
		t.Errorf("after deleting its Task: %+v", got)
	}
}

func TestRowsThatCannotBeMovedSayWhyAndStayPut(t *testing.T) {
	c := New()
	c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Weekdays: []time.Weekday{time.Friday}})
	c.AddTask(today, "Pack", "passport")
	c.AddTask(today, "Buy milk")
	c.Tick(row(t, c, today, "Buy milk"), today)

	before := c.Clone()
	for title, want := range map[string]error{
		"Study":    ErrRecurringMove,
		"passport": ErrSubtaskMove,
		"Buy milk": ErrDoneMove,
	} {
		r := row(t, c, today, title)
		if got := r.Movable(); got != want {
			t.Errorf("%s: Movable = %v, want %v", title, got, want)
		}
		if got := c.Move(r, d("2026-10-16")); got != want {
			t.Errorf("%s: Move = %v, want %v", title, got, want)
		}
	}
	if !reflect.DeepEqual(c, before) {
		t.Error("a refused Move changed the Calendar")
	}
}

func TestMovedTaskTakesItsSubtasksWithIt(t *testing.T) {
	c := New()
	c.AddTask(today, "Pack", "passport", "charger")
	c.Tick(row(t, c, today, "passport"), today)

	if err := c.Move(row(t, c, today, "Pack"), d("2026-10-16")); err != nil {
		t.Fatal(err)
	}
	if n := len(c.Day(today, today).Rows); n != 0 {
		t.Errorf("%d Rows left on the old day", n)
	}
	want := []line{{TaskRow, Open, "[ ] Pack 1/2"}, {SubtaskRow, Done, "[x] passport"}, {SubtaskRow, Open, "[ ] charger"}}
	if got := lines(c.Day(d("2026-10-16"), today)); !reflect.DeepEqual(got, want) {
		t.Errorf("new day: got %+v, want %+v", got, want)
	}
}
