package calendar

import (
	"errors"
	"slices"
)

// Why a Row can't be Moved.
var (
	ErrSubtaskMove   = errors.New("a subtask moves with its task")
	ErrRecurringMove = errors.New("a recurring event can't be moved")
	ErrDoneMove      = errors.New("a done task can't be moved")
)

// Tick ticks or unticks the Task or Subtask a Row stands for. A Task ticked
// today is Done today; a Subtask's tick leaves its Task as it is. An Event
// has no tick.
func (c *Calendar) Tick(r Row, today Date) {
	switch r.Kind {
	case TaskRow:
		c.toggleTask(r.Task.ID, today)
	case SubtaskRow:
		c.toggleSubtask(r.Task.ID, r.Sub)
	}
}

// Occurrence reports whether the Row is one Occurrence of a Recurring Event,
// which can be taken away with Skip as well as with Delete.
func (r Row) Occurrence() bool { return r.Kind == EventRow && r.Event.Recurring() }

// Skip removes the Occurrence a Row stands for and leaves the rest of its
// Recurring Event.
func (c *Calendar) Skip(r Row) {
	if !r.Occurrence() {
		return
	}
	if e := c.event(r.Event.ID); e != nil && e.Recurring() && e.OccursOn(r.Date) {
		e.Skips = append(e.Skips, r.Date)
	}
}

// Delete removes what a Row stands for: a Subtask on its own, a Task with its
// Subtasks, or an Event with every Occurrence.
func (c *Calendar) Delete(r Row) {
	switch r.Kind {
	case EventRow:
		c.Events = slices.DeleteFunc(c.Events, func(e Event) bool { return e.ID == r.Event.ID })
	case TaskRow:
		c.Tasks = slices.DeleteFunc(c.Tasks, func(t Task) bool { return t.ID == r.Task.ID })
	case SubtaskRow:
		if c.subtask(r.Task.ID, r.Sub) != nil {
			t := c.task(r.Task.ID)
			t.Subtasks = slices.Delete(t.Subtasks, r.Sub, r.Sub+1)
		}
	}
}

// Movable returns nil if the Row can be Moved, and otherwise why not: only
// an unfinished Task and a One-off Event can.
func (r Row) Movable() error {
	switch {
	case r.Kind == SubtaskRow:
		return ErrSubtaskMove
	case r.Occurrence():
		return ErrRecurringMove
	case r.Kind == TaskRow && r.State == Done:
		return ErrDoneMove
	}
	return nil
}

// Move attaches the Task or One-off Event a Row stands for to another date.
// A Task takes its Subtasks with it and is no longer Overdue. A Row that is
// not Movable stays where it is, and Move returns why.
func (c *Calendar) Move(r Row, to Date) error {
	if err := r.Movable(); err != nil {
		return err
	}
	if r.Kind == TaskRow {
		if t := c.task(r.Task.ID); t != nil {
			t.Date = to
		}
	} else if e := c.event(r.Event.ID); e != nil {
		e.Date = to
	}
	return nil
}

// CycleColor gives the Event a Row stands for the next Marker colour.
func (c *Calendar) CycleColor(r Row) {
	if e := c.event(r.Event.ID); r.Kind == EventRow && e != nil {
		e.Color = (e.Color + 1) % Colors
	}
}
