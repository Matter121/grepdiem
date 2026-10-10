package calendar

// RowKind says what a Row stands for.
type RowKind int

const (
	EventRow RowKind = iota
	TodoRow
)

// TodoState is how a Todo stands on the day it is listed.
type TodoState int

const (
	Open    TodoState = iota
	Overdue           // unfinished, listed under today after its own date has passed
	Done              // ticked, listed on the day it was ticked
)

// Row is one line in a day's listing, standing for one Event or one Todo.
// Its text is plain: colour and layout are up to whoever shows it.
type Row struct {
	Kind  RowKind
	Event Event     // what an EventRow stands for
	Todo  Todo      // what a TodoRow stands for
	State TodoState // of a TodoRow; an EventRow is always Open
}

func (r Row) Title() string {
	if r.Kind == TodoRow {
		return r.Todo.Title
	}
	return r.Event.Title
}

// Time is an Event's time of day, "13:00" or "13:00–15:00". It is empty for
// an All-day Event and for a Todo.
func (r Row) Time() string { return TimeRange(r.Event.Start, r.Event.End) }

// Mark is a Todo's tick box, "[x]" once Done. It is empty for an Event.
func (r Row) Mark() string {
	switch {
	case r.Kind != TodoRow:
		return ""
	case r.State == Done:
		return "[x]"
	}
	return "[ ]"
}

// Note says that a Todo is Overdue and from which date. It is empty for any
// other Row.
func (r Row) Note() string {
	if r.State != Overdue {
		return ""
	}
	return "overdue from " + r.Todo.Date.Format("Mon 2 Jan")
}
