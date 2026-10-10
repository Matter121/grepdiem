package main

import (
	"strings"
	"testing"

	"github.com/Matter121/TUICalendar/calendar"
)

func TestTodayPrintsEventsThenTodos(t *testing.T) {
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
	c.AddTodo(today, "Buy milk")
	c.AddTodo(earlier, "Call landlord")
	ticked := c.AddTodo(today, "Water plants")
	c.ToggleTodo(ticked.ID, today)

	var out strings.Builder
	printToday(&out, c, today)
	want := `Fri 9 Oct 2026 · week 41

Events
  all day      Birthday
  13:00–15:00  Lecture

Todos
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
