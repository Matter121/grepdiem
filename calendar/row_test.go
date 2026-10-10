package calendar

import (
	"testing"
	"time"
)

func TestDayListsEventRowsThenTodoRowsWithTheirStateAndText(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	c.AddEvent(today, EventSpec{Title: "Lecture", Start: "13:00", End: "15:00"})
	c.AddEvent(d("2026-10-02"), EventSpec{Title: "Gym", Start: "07:30", Weekdays: []time.Weekday{time.Friday}})
	c.AddEvent(today, EventSpec{Title: "Birthday"})
	c.AddTodo(today, "Buy milk")
	c.AddTodo(d("2026-10-06"), "Call landlord")
	ticked := c.AddTodo(today, "Water plants")
	c.ToggleTodo(ticked.ID, today)

	type row struct {
		kind              RowKind
		state             TodoState
		title, time, mark string
		note              string
	}
	want := []row{
		{EventRow, Open, "Birthday", "", "", ""},
		{EventRow, Open, "Gym", "07:30", "", ""},
		{EventRow, Open, "Lecture", "13:00–15:00", "", ""},
		{TodoRow, Overdue, "Call landlord", "", "[ ]", "from Tue 6 Oct"},
		{TodoRow, Open, "Buy milk", "", "[ ]", ""},
		{TodoRow, Done, "Water plants", "", "[x]", ""},
	}

	var got []row
	for _, r := range c.Day(today, today).Rows {
		got = append(got, row{r.Kind, r.State, r.Title(), r.Time(), r.Mark(), r.Note()})
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestRowStandsForItsEventOrTodo(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	e := c.AddEvent(today, EventSpec{Title: "Lecture"})
	todo := c.AddTodo(today, "Buy milk")

	rows := c.Day(today, today).Rows
	if len(rows) != 2 || rows[0].Event.ID != e.ID || rows[1].Todo.ID != todo.ID {
		t.Errorf("rows %+v, want the event %d then the todo %d", rows, e.ID, todo.ID)
	}
}

func TestDayWithNothingOnItHasNoRows(t *testing.T) {
	c := New()
	c.AddTodo(d("2026-10-12"), "Hand in essay")
	if rows := c.Day(d("2026-10-09"), d("2026-10-09")).Rows; len(rows) != 0 {
		t.Errorf("rows %+v, want none", rows)
	}
}
