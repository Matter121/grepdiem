// Package calendar holds Events and Tasks and answers what is on a given day.
// It does no I/O and never reads the system clock.
package calendar

import (
	"encoding/json"
	"slices"
	"sort"
	"time"
)

// Colors is the number of Marker colours an Event can have.
const Colors = 6

// EventSpec is the part of an Event that Quick-add describes.
type EventSpec struct {
	Title    string
	Start    Clock
	End      Clock
	Weekdays []time.Weekday
	Until    Date
}

// Recurring reports whether the spec describes a Recurring Event.
func (s EventSpec) Recurring() bool { return len(s.Weekdays) > 0 }

// Event is something that happens on a date. With Weekdays set it is a
// Recurring Event whose Occurrences are those weekdays from Date to Until.
type Event struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Date     Date     `json:"date"`
	Start    Clock    `json:"start,omitempty"`
	End      Clock    `json:"end,omitempty"`
	Weekdays Weekdays `json:"weekdays,omitempty"`
	Until    Date     `json:"until,omitzero"`
	Skips    []Date   `json:"skips,omitempty"`
	Color    int      `json:"color"`
}

func (e Event) Recurring() bool { return len(e.Weekdays) > 0 }

// Spec returns the Quick-add part of the Event.
func (e Event) Spec() EventSpec {
	return EventSpec{e.Title, e.Start, e.End, e.Weekdays, e.Until}
}

// OccursOn reports whether the Event has an Occurrence on d.
func (e Event) OccursOn(d Date) bool {
	if !e.Recurring() {
		return d == e.Date
	}
	if d.Before(e.Date) || (!e.Until.IsZero() && d.After(e.Until)) {
		return false
	}
	return slices.Contains(e.Weekdays, d.Weekday()) && !slices.Contains(e.Skips, d)
}

// Calendar is everything the user has entered.
type Calendar struct {
	Version int     `json:"version"`
	NextID  int     `json:"next_id"`
	Events  []Event `json:"events"`
	Tasks   []Task  `json:"tasks"`
}

// version is the data format written. Version 1 kept its Tasks under "todos".
const version = 2

func New() *Calendar {
	return &Calendar{Version: version, NextID: 1}
}

// UnmarshalJSON reads the current format and version 1, and leaves the
// Calendar as the current one either way.
func (c *Calendar) UnmarshalJSON(data []byte) error {
	type plain Calendar
	file := struct {
		*plain
		Todos []Task `json:"todos"`
	}{plain: (*plain)(c)}
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	if c.Version < 2 {
		c.Tasks = append(c.Tasks, file.Todos...)
	}
	c.Version = version
	return nil
}

func (c *Calendar) newID() int {
	id := c.NextID
	c.NextID++
	return id
}

// Day is what one date looks like.
type Day struct {
	Date Date
	// Rows are the Events, All-day first and then by start time, followed by
	// the Tasks: Overdue, then open, then Done. An unfinished Task is followed
	// by its Subtasks.
	Rows []Row
}

// WithoutSubtasks returns the Rows that stand for an Event or a Task.
func (d Day) WithoutSubtasks() []Row {
	return slices.DeleteFunc(slices.Clone(d.Rows), func(r Row) bool { return r.Kind == SubtaskRow })
}

// Day returns what is on d, given which date is today.
func (c *Calendar) Day(d, today Date) Day {
	var events []Event
	for _, e := range c.Events {
		if e.OccursOn(d) {
			events = append(events, e)
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Start < events[j].Start
	})
	day := Day{Date: d}
	for _, e := range events {
		day.Rows = append(day.Rows, Row{Kind: EventRow, Event: e})
	}
	day.Rows = append(day.Rows, c.taskRows(d, today)...)
	for i := range day.Rows {
		day.Rows[i].Date = d
	}
	return day
}

// AddEvent creates an Event starting on date and gives it the next colour.
func (c *Calendar) AddEvent(date Date, spec EventSpec) Event {
	e := Event{ID: c.newID(), Date: date, Color: len(c.Events) % Colors}
	e.apply(spec)
	c.Events = append(c.Events, e)
	return e
}

func (e *Event) apply(spec EventSpec) {
	e.Title, e.Start, e.End = spec.Title, spec.Start, spec.End
	e.Weekdays, e.Until = slices.Clone(Weekdays(spec.Weekdays)), spec.Until
	if !e.Recurring() {
		e.Until, e.Skips = Date{}, nil
	}
}

func (c *Calendar) event(id int) *Event {
	for i := range c.Events {
		if c.Events[i].ID == id {
			return &c.Events[i]
		}
	}
	return nil
}

// UpdateEvent replaces what Quick-add describes. For a Recurring Event this
// changes every Occurrence; removing the weekdays leaves one Event on Date.
func (c *Calendar) UpdateEvent(id int, spec EventSpec) {
	if e := c.event(id); e != nil {
		e.apply(spec)
	}
}

// First returns the Event's first Occurrence, or the zero Date if it has none.
func (e Event) First() Date {
	if !e.Recurring() {
		return e.Date
	}
	for d := e.Date; e.Until.IsZero() || !d.After(e.Until); d = d.AddDays(1) {
		if e.OccursOn(d) {
			return d
		}
	}
	return Date{}
}

// Clone returns a copy that shares nothing with c.
func (c *Calendar) Clone() *Calendar {
	cp := *c
	cp.Tasks = slices.Clone(c.Tasks)
	for i := range cp.Tasks {
		cp.Tasks[i].Subtasks = slices.Clone(cp.Tasks[i].Subtasks)
	}
	cp.Events = slices.Clone(c.Events)
	for i := range cp.Events {
		cp.Events[i].Weekdays = slices.Clone(cp.Events[i].Weekdays)
		cp.Events[i].Skips = slices.Clone(cp.Events[i].Skips)
	}
	return &cp
}
