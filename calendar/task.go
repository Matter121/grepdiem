package calendar

import (
	"slices"
	"sort"
)

// Task is something to be done and ticked off, attached to a date.
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Date  Date   `json:"date"`
	Done  Date   `json:"done,omitzero"` // the day it was ticked
}

// taskRows lists Overdue Tasks first, then open ones, then Done ones.
func (c *Calendar) taskRows(d, today Date) []Row {
	var overdue, open, done []Row
	for _, t := range c.Tasks {
		switch {
		case !t.Done.IsZero():
			if t.Done == d {
				done = append(done, Row{Kind: TaskRow, Task: t, State: Done})
			}
		case d == today && t.Date.Before(today):
			overdue = append(overdue, Row{Kind: TaskRow, Task: t, State: Overdue})
		case t.Date == d && !d.Before(today):
			open = append(open, Row{Kind: TaskRow, Task: t})
		}
	}
	sort.SliceStable(overdue, func(i, j int) bool {
		return overdue[i].Task.Date.Before(overdue[j].Task.Date)
	})
	return append(append(overdue, open...), done...)
}

func (c *Calendar) task(id int) *Task {
	for i := range c.Tasks {
		if c.Tasks[i].ID == id {
			return &c.Tasks[i]
		}
	}
	return nil
}

// AddTask attaches a new Task to date.
func (c *Calendar) AddTask(date Date, title string) Task {
	t := Task{ID: c.newID(), Title: title, Date: date}
	c.Tasks = append(c.Tasks, t)
	return t
}

// ToggleTask ticks an open Task as Done today, or reopens a Done one.
func (c *Calendar) ToggleTask(id int, today Date) {
	if t := c.task(id); t != nil {
		if t.Done.IsZero() {
			t.Done = today
		} else {
			t.Done = Date{}
		}
	}
}

// RenameTask changes a Task's title.
func (c *Calendar) RenameTask(id int, title string) {
	if t := c.task(id); t != nil {
		t.Title = title
	}
}

// MoveTask attaches a Task to another date.
func (c *Calendar) MoveTask(id int, date Date) {
	if t := c.task(id); t != nil {
		t.Date = date
	}
}

func (c *Calendar) DeleteTask(id int) {
	c.Tasks = slices.DeleteFunc(c.Tasks, func(t Task) bool { return t.ID == id })
}
