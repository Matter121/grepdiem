package calendar

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Task is something to be done and ticked off, attached to a date.
type Task struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Date     Date      `json:"date"`
	Done     Date      `json:"done,omitzero"` // the day it was ticked
	Subtasks []Subtask `json:"subtasks,omitempty"`
}

// Subtask is one step of a Task, ticked on its own. It has no date: it goes
// wherever its Task goes.
type Subtask struct {
	Title string `json:"title"`
	Done  bool   `json:"done,omitempty"`
}

// TaskSpec is the part of a Task that Quick-add describes.
type TaskSpec struct {
	Title    string
	Subtasks []string
}

// Spec returns the Quick-add part of the Task.
func (t Task) Spec() TaskSpec {
	spec := TaskSpec{Title: t.Title}
	for _, s := range t.Subtasks {
		spec.Subtasks = append(spec.Subtasks, s.Title)
	}
	return spec
}

// ParseTask reads a Quick-add line for a Task.
//
// The first word "subtasks" (or "s") that has a word before it and text
// after it ends the title; what follows are the Subtasks, separated by
// commas. A title given in double quotes is kept exactly as typed.
func ParseTask(line string) TaskSpec {
	title, words, quoted := splitQuoted(line)
	first := 1 // the first word is always title
	if quoted {
		first = 0
	}
	for i := first; i < len(words); i++ {
		if !strings.EqualFold(words[i], "subtasks") && !strings.EqualFold(words[i], "s") {
			continue
		}
		var subtasks []string
		for _, part := range strings.Split(strings.Join(words[i+1:], " "), ",") {
			if part = strings.TrimSpace(part); part != "" {
				subtasks = append(subtasks, part)
			}
		}
		if len(subtasks) > 0 {
			words = words[:i]
			return TaskSpec{strings.TrimSpace(title + " " + strings.Join(words, " ")), subtasks}
		}
		break
	}
	return TaskSpec{Title: strings.TrimSpace(title + " " + strings.Join(words, " "))}
}

// TaskLine returns the Quick-add line that recreates spec.
func TaskLine(spec TaskSpec) string {
	tail := ""
	if len(spec.Subtasks) > 0 {
		tail = " s " + strings.Join(spec.Subtasks, ", ")
	}
	line := spec.Title + tail
	// Quote a title that would otherwise be read as having Subtasks.
	if got := ParseTask(line); got.Title != spec.Title || !slices.Equal(got.Subtasks, spec.Subtasks) {
		line = `"` + spec.Title + `"` + tail
	}
	return line
}

// DescribeTask says in words how a Quick-add line for a Task was understood.
func DescribeTask(spec TaskSpec) string {
	switch n := len(spec.Subtasks); n {
	case 0:
		return spec.Title
	case 1:
		return spec.Title + " · 1 subtask: " + spec.Subtasks[0]
	default:
		return fmt.Sprintf("%s · %d subtasks: %s", spec.Title, n, strings.Join(spec.Subtasks, ", "))
	}
}

// taskRows lists Overdue Tasks first, then open ones, then Done ones. An
// unfinished Task is followed by its Subtasks; a Done one is a single Row.
func (c *Calendar) taskRows(d, today Date) []Row {
	var overdue, open, done []Row
	for _, t := range c.Tasks {
		switch {
		case !t.Done.IsZero():
			if t.Done == d {
				done = append(done, Row{Kind: TaskRow, Task: t, State: Done})
			}
		case d == today && t.Date.Before(today):
			overdue = append(append(overdue, Row{Kind: TaskRow, Task: t, State: Overdue}), subtaskRows(t)...)
		case t.Date == d && !d.Before(today):
			open = append(append(open, Row{Kind: TaskRow, Task: t}), subtaskRows(t)...)
		}
	}
	sort.SliceStable(overdue, func(i, j int) bool {
		return overdue[i].Task.Date.Before(overdue[j].Task.Date)
	})
	return append(append(overdue, open...), done...)
}

func subtaskRows(t Task) []Row {
	rows := make([]Row, len(t.Subtasks))
	for i, s := range t.Subtasks {
		rows[i] = Row{Kind: SubtaskRow, Task: t, Sub: i}
		if s.Done {
			rows[i].State = Done
		}
	}
	return rows
}

func (c *Calendar) task(id int) *Task {
	for i := range c.Tasks {
		if c.Tasks[i].ID == id {
			return &c.Tasks[i]
		}
	}
	return nil
}

func (c *Calendar) subtask(id, sub int) *Subtask {
	if t := c.task(id); t != nil && sub >= 0 && sub < len(t.Subtasks) {
		return &t.Subtasks[sub]
	}
	return nil
}

// AddTask attaches a new Task to date.
func (c *Calendar) AddTask(date Date, title string, subtasks ...string) Task {
	t := Task{ID: c.newID(), Title: title, Date: date}
	for _, s := range subtasks {
		t.Subtasks = append(t.Subtasks, Subtask{Title: s})
	}
	c.Tasks = append(c.Tasks, t)
	return t
}

// toggleTask ticks an open Task as Done today, or reopens a Done one.
func (c *Calendar) toggleTask(id int, today Date) {
	if t := c.task(id); t != nil {
		if t.Done.IsZero() {
			t.Done = today
		} else {
			t.Done = Date{}
		}
	}
}

// UpdateTask replaces what Quick-add describes. A Subtask whose title is
// still there keeps its tick; any other is new and unticked.
func (c *Calendar) UpdateTask(id int, title string, subtasks ...string) {
	t := c.task(id)
	if t == nil {
		return
	}
	old := t.Subtasks
	t.Title, t.Subtasks = title, nil
	for _, s := range subtasks {
		next := Subtask{Title: s}
		if i := slices.IndexFunc(old, func(o Subtask) bool { return o.Title == s }); i >= 0 {
			next.Done = old[i].Done
			old = slices.Delete(old, i, i+1)
		}
		t.Subtasks = append(t.Subtasks, next)
	}
}

// toggleSubtask ticks or unticks a Subtask. Its Task stays as it is.
func (c *Calendar) toggleSubtask(id, sub int) {
	if s := c.subtask(id, sub); s != nil {
		s.Done = !s.Done
	}
}

// RenameSubtask changes a Subtask's title and keeps its tick.
func (c *Calendar) RenameSubtask(id, sub int, title string) {
	if s := c.subtask(id, sub); s != nil {
		s.Title = title
	}
}
