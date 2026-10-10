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
	c.Skip(row(t, c, d("2026-10-13"), "Study"))
	task := c.AddTask(d("2026-10-06"), "Call landlord")
	c.toggleTask(task.ID, d("2026-10-09"))
	c.AddTask(d("2026-10-09"), "Buy milk")

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
		t.Errorf("open Task should have no done field, found %d done fields", n)
	}

	back := New()
	if err := json.Unmarshal(data, back); err != nil {
		t.Fatal(err)
	}
	today := d("2026-10-09")
	for _, date := range []string{"2026-10-08", "2026-10-09", "2026-10-13", "2026-10-15"} {
		a, b := c.Day(d(date), today), back.Day(d(date), today)
		if len(a.Rows) != len(b.Rows) || !sameShown(tasksShown(a), tasksShown(b)) {
			t.Errorf("%s differs after JSON: %+v vs %+v", date, a, b)
		}
	}
	// IDs keep counting from where they were.
	if next := back.AddTask(today, "New"); next.ID == task.ID || next.ID == e.ID {
		t.Errorf("reused ID %d", next.ID)
	}
}

func TestVersion1FileIsReadWithItsTodosAsTasks(t *testing.T) {
	old := `{"version":1,"next_id":3,"events":[],"todos":[
		{"id":1,"title":"Call landlord","date":"2026-10-06","done":"2026-10-09"},
		{"id":2,"title":"Buy milk","date":"2026-10-09"}]}`
	c := New()
	if err := json.Unmarshal([]byte(old), c); err != nil {
		t.Fatal(err)
	}
	today := d("2026-10-09")
	want := []shown{{"Buy milk", Open}, {"Call landlord", Done}}
	if got := tasksShown(c.Day(today, today)); !sameShown(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if s := string(data); !strings.Contains(s, `"version":2`) || !strings.Contains(s, `"tasks":[`) || strings.Contains(s, `"todos"`) {
		t.Errorf("should be written as version 2 with tasks:\n%s", s)
	}
}
