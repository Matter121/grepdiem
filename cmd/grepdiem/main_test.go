package main

import (
	"strings"
	"testing"

	"github.com/Matter121/grepdiem/calendar"
)

func TestTodayPrintsEventsThenTasks(t *testing.T) {
	today, err := calendar.ParseDate("2026-10-09")
	if err != nil {
		t.Fatal(err)
	}
	earlier, err := calendar.ParseDate("2026-10-06")
	if err != nil {
		t.Fatal(err)
	}
	c := calendar.New()
	c.AddEvent(today, calendar.EventSpec{Title: "Lecture", Start: "13:00", End: "15:00"})
	c.AddEvent(today, calendar.EventSpec{Title: "Birthday"})
	c.AddTask(today, "Buy milk")
	c.AddTask(earlier, "Call landlord")
	ticked := c.AddTask(today, "Water plants")
	c.ToggleTask(ticked.ID, today)

	var out strings.Builder
	printToday(&out, c, today)
	want := `Fri 9 Oct 2026 · week 41

Events
  all day      Birthday
  13:00–15:00  Lecture

Tasks
  [ ] Call landlord (overdue from Tue 6 Oct)
  [ ] Buy milk
  [x] Water plants
`
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestTodayWithNothingOnItSaysSo(t *testing.T) {
	today, err := calendar.ParseDate("2026-10-09")
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	printToday(&out, calendar.New(), today)
	if want := "Fri 9 Oct 2026 · week 41\n\nNothing today.\n"; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
