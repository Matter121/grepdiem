package ui

import "github.com/Matter121/grepdiem/calendar"

// listed returns the Rows a day lists in a view: all of them in the Compact
// view, and no Subtasks in the Full view, which has no room for them.
func listed(day calendar.Day, full bool) []calendar.Row {
	if full {
		return day.WithoutSubtasks()
	}
	return day.Rows
}

// listing is the selected day's Rows as the view in use lists them, and the
// cursor on one of them. Both views draw from it and keys act on its current
// Row, so the Row that is marked is the Row that is acted on.
//
// The cursor is a place in the list, not a Row: when a change reorders the
// Rows it stays where it was, so after ticking a Task it is on the next one.
type listing struct {
	rows    []calendar.Row
	cursor  int  // always a valid index into rows, unless there are none
	focused bool // the day is open, so the cursor shows and keys act on its Row
}

func (m *Model) listing() listing {
	rows := listed(m.store.Calendar().Day(m.selected, m.today()), m.full)
	return listing{rows, max(0, min(m.cursor, len(rows)-1)), m.focus == focusPanel}
}

// current returns the Row keys act on, if the day is open and has any.
func (l listing) current() (calendar.Row, bool) {
	if !l.focused || len(l.rows) == 0 {
		return calendar.Row{}, false
	}
	return l.rows[l.cursor], true
}

// mark is the index of the Row to draw the cursor on, or -1 for none.
func (l listing) mark() int {
	if _, ok := l.current(); !ok {
		return -1
	}
	return l.cursor
}

// step returns where the cursor is after moving by rows up or down.
func (l listing) step(by int) int {
	return max(0, min(l.cursor+by, len(l.rows)-1))
}

// find returns where the cursor has to be to stay on r's Event or Task. A
// Subtask's place is its Task's. If neither is listed the cursor stays.
func (l listing) find(r calendar.Row) int {
	for i, o := range l.rows {
		if o.Kind != calendar.SubtaskRow && o.Event.ID == r.Event.ID && o.Task.ID == r.Task.ID {
			return i
		}
	}
	return l.cursor
}
