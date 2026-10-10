package calendar

import (
	"slices"
	"sort"
)

// Todo is something to be done and ticked off, attached to a date.
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Date  Date   `json:"date"`
	Done  Date   `json:"done,omitzero"` // the day it was ticked
}

// todoRows lists Overdue Todos first, then open ones, then Done ones.
func (c *Calendar) todoRows(d, today Date) []Row {
	var overdue, open, done []Row
	for _, t := range c.Todos {
		switch {
		case !t.Done.IsZero():
			if t.Done == d {
				done = append(done, Row{Kind: TodoRow, Todo: t, State: Done})
			}
		case d == today && t.Date.Before(today):
			overdue = append(overdue, Row{Kind: TodoRow, Todo: t, State: Overdue})
		case t.Date == d && !d.Before(today):
			open = append(open, Row{Kind: TodoRow, Todo: t})
		}
	}
	sort.SliceStable(overdue, func(i, j int) bool {
		return overdue[i].Todo.Date.Before(overdue[j].Todo.Date)
	})
	return append(append(overdue, open...), done...)
}

func (c *Calendar) todo(id int) *Todo {
	for i := range c.Todos {
		if c.Todos[i].ID == id {
			return &c.Todos[i]
		}
	}
	return nil
}

// AddTodo attaches a new Todo to date.
func (c *Calendar) AddTodo(date Date, title string) Todo {
	t := Todo{ID: c.newID(), Title: title, Date: date}
	c.Todos = append(c.Todos, t)
	return t
}

// ToggleTodo ticks an open Todo as Done today, or reopens a Done one.
func (c *Calendar) ToggleTodo(id int, today Date) {
	if t := c.todo(id); t != nil {
		if t.Done.IsZero() {
			t.Done = today
		} else {
			t.Done = Date{}
		}
	}
}

// RenameTodo changes a Todo's title.
func (c *Calendar) RenameTodo(id int, title string) {
	if t := c.todo(id); t != nil {
		t.Title = title
	}
}

// MoveTodo attaches a Todo to another date.
func (c *Calendar) MoveTodo(id int, date Date) {
	if t := c.todo(id); t != nil {
		t.Date = date
	}
}

func (c *Calendar) DeleteTodo(id int) {
	c.Todos = slices.DeleteFunc(c.Todos, func(t Todo) bool { return t.ID == id })
}
