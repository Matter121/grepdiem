package calendar

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCalendarSurvivesJSONAndReadsNaturally(t *testing.T) {
	c := New()
	e := c.AddEvent(d("2026-10-06"), EventSpec{Title: "Study", Start: "13:00",
		Weekdays: []time.Weekday{time.Tuesday, time.Thursday}, Until: d("2026-12-13")})
	c.Skip(e.ID, d("2026-10-13"))
	todo := c.AddTodo(d("2026-10-06"), "Call landlord")
	c.ToggleTodo(todo.ID, d("2026-10-09"))
	c.AddTodo(d("2026-10-09"), "Buy milk")

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"weekdays":["tue","thu"]`, `"until":"2026-12-13"`, `"skips":["2026-10-13"]`, `"done":"2026-10-09"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("JSON lacks %s:\n%s", want, data)
		}
	}
	if n := strings.Count(string(data), `"done"`); n != 1 {
		t.Errorf("open Todo should have no done field, found %d done fields", n)
	}

	back := New()
	if err := json.Unmarshal(data, back); err != nil {
		t.Fatal(err)
	}
	today := d("2026-10-09")
	for _, date := range []string{"2026-10-08", "2026-10-09", "2026-10-13", "2026-10-15"} {
		a, b := c.Day(d(date), today), back.Day(d(date), today)
		if len(a.Rows) != len(b.Rows) || !sameShown(todosShown(a), todosShown(b)) {
			t.Errorf("%s differs after JSON: %+v vs %+v", date, a, b)
		}
	}
	// IDs keep counting from where they were.
	if next := back.AddTodo(today, "New"); next.ID == todo.ID || next.ID == e.ID {
		t.Errorf("reused ID %d", next.ID)
	}
}
