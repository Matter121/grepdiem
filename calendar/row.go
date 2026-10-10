package calendar

// RowKind says what a Row stands for.
type RowKind int

const (
	EventRow RowKind = iota
	TaskRow
)

// TaskState is how a Task stands on the day it is listed.
type TaskState int

const (
	Open    TaskState = iota
	Overdue           // unfinished, listed under today after its own date has passed
	Done              // ticked, listed on the day it was ticked
)

// Row is one line in a day's listing, standing for one Event or one Task.
// Its text is plain: colour and layout are up to whoever shows it.
type Row struct {
	Kind  RowKind
	Event Event     // what an EventRow stands for
	Task  Task      // what a TaskRow stands for
	State TaskState // of a TaskRow; an EventRow is always Open
}

func (r Row) Title() string {
	if r.Kind == TaskRow {
		return r.Task.Title
	}
	return r.Event.Title
}

// Time is an Event's time of day, "13:00" or "13:00–15:00". It is empty for
// an All-day Event and for a Task.
func (r Row) Time() string { return TimeRange(r.Event.Start, r.Event.End) }

// Mark is a Task's tick box, "[x]" once Done. It is empty for an Event.
func (r Row) Mark() string {
	switch {
	case r.Kind != TaskRow:
		return ""
	case r.State == Done:
		return "[x]"
	}
	return "[ ]"
}

// Note says that a Task is Overdue and from which date. It is empty for any
// other Row.
func (r Row) Note() string {
	if r.State != Overdue {
		return ""
	}
	return "overdue from " + r.Task.Date.Format("Mon 2 Jan")
}
