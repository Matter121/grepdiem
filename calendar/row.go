package calendar

import "fmt"

// RowKind says what a Row stands for.
type RowKind int

const (
	EventRow RowKind = iota
	TaskRow
	SubtaskRow
)

// TaskState is how a Task stands on the day it is listed. A Subtask is only
// ever Open or Done.
type TaskState int

const (
	Open    TaskState = iota
	Overdue           // unfinished, listed under today after its own date has passed
	Done              // ticked, listed on the day it was ticked
)

// Row is one line in a day's listing, standing for one Event, one Task or
// one Subtask. Its text is plain: colour and layout are up to whoever shows it.
type Row struct {
	Kind  RowKind
	Event Event     // what an EventRow stands for
	Task  Task      // what a TaskRow stands for, or the Task a SubtaskRow belongs to
	Sub   int       // which of the Task's Subtasks a SubtaskRow stands for
	State TaskState // of a TaskRow or SubtaskRow; an EventRow is always Open
}

func (r Row) Title() string {
	switch r.Kind {
	case TaskRow:
		return r.Task.Title
	case SubtaskRow:
		return r.Task.Subtasks[r.Sub].Title
	}
	return r.Event.Title
}

// Count is how many of a Task's Subtasks are ticked, "1/3". It is empty for
// a Task without Subtasks and for any other Row.
func (r Row) Count() string {
	if r.Kind != TaskRow || len(r.Task.Subtasks) == 0 {
		return ""
	}
	ticked := 0
	for _, s := range r.Task.Subtasks {
		if s.Done {
			ticked++
		}
	}
	return fmt.Sprintf("%d/%d", ticked, len(r.Task.Subtasks))
}

// Time is an Event's time of day, "13:00" or "13:00–15:00". It is empty for
// an All-day Event and for a Task.
func (r Row) Time() string { return TimeRange(r.Event.Start, r.Event.End) }

// Mark is the tick box of a Task or Subtask, "[x]" once Done. It is empty for
// an Event.
func (r Row) Mark() string {
	switch {
	case r.Kind == EventRow:
		return ""
	case r.State == Done:
		return "[x]"
	}
	return "[ ]"
}

// Note says that a Task is Overdue and from which date. It is empty for any
// other Row.
func (r Row) Note() string {
	if r.Kind != TaskRow || r.State != Overdue {
		return ""
	}
	return "overdue from " + r.Task.Date.Format("Mon 2 Jan")
}
