package calendar

import "testing"

type shown struct {
	title   string
	overdue bool
	done    bool
}

func todosShown(day Day) []shown {
	var out []shown
	for _, t := range day.Todos {
		out = append(out, shown{t.Title, t.Overdue, !t.Done.IsZero()})
	}
	return out
}

func sameShown(a, b []shown) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTodoShowsOnItsDate(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	c.AddTodo(today, "Buy milk")
	c.AddTodo(d("2026-10-12"), "Hand in essay")

	if got := todosShown(c.Day(today, today)); !sameShown(got, []shown{{"Buy milk", false, false}}) {
		t.Errorf("today: %v", got)
	}
	if got := todosShown(c.Day(d("2026-10-12"), today)); !sameShown(got, []shown{{"Hand in essay", false, false}}) {
		t.Errorf("future day: %v", got)
	}
}

func TestUnfinishedTodoRollsOverToTodayAsOverdue(t *testing.T) {
	c := New()
	c.AddTodo(d("2026-10-06"), "Call landlord")
	c.AddTodo(d("2026-10-09"), "Buy milk")
	today := d("2026-10-09")

	want := []shown{{"Call landlord", true, false}, {"Buy milk", false, false}}
	if got := todosShown(c.Day(today, today)); !sameShown(got, want) {
		t.Errorf("today: got %v, want %v", got, want)
	}
	if got := todosShown(c.Day(d("2026-10-06"), today)); len(got) != 0 {
		t.Errorf("the day it was attached to should no longer show it: %v", got)
	}
}

func TestTickedOverdueTodoIsDoneOnTheDayItWasTicked(t *testing.T) {
	c := New()
	todo := c.AddTodo(d("2026-10-06"), "Call landlord")
	today := d("2026-10-09")
	c.ToggleTodo(todo.ID, today)

	if got := todosShown(c.Day(today, today)); !sameShown(got, []shown{{"Call landlord", false, true}}) {
		t.Errorf("day ticked: %v", got)
	}
	// Still visible there the next day, and not Overdue again.
	tomorrow := d("2026-10-10")
	if got := todosShown(c.Day(today, tomorrow)); !sameShown(got, []shown{{"Call landlord", false, true}}) {
		t.Errorf("day ticked, seen tomorrow: %v", got)
	}
	if got := todosShown(c.Day(tomorrow, tomorrow)); len(got) != 0 {
		t.Errorf("tomorrow: %v", got)
	}
}

func TestUntickingMakesATodoOpenAgain(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	todo := c.AddTodo(today, "Buy milk")
	c.ToggleTodo(todo.ID, today)
	c.ToggleTodo(todo.ID, today)
	if got := todosShown(c.Day(today, today)); !sameShown(got, []shown{{"Buy milk", false, false}}) {
		t.Errorf("got %v", got)
	}
}

func TestOpenTodosComeBeforeDoneOnes(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	first := c.AddTodo(today, "First")
	c.AddTodo(today, "Second")
	c.ToggleTodo(first.ID, today)

	want := []shown{{"Second", false, false}, {"First", false, true}}
	if got := todosShown(c.Day(today, today)); !sameShown(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMovedOverdueTodoIsAnOrdinaryTodoOnItsNewDay(t *testing.T) {
	c := New()
	todo := c.AddTodo(d("2026-10-06"), "Call landlord")
	today := d("2026-10-09")
	c.MoveTodo(todo.ID, d("2026-10-16"))

	if got := todosShown(c.Day(today, today)); len(got) != 0 {
		t.Errorf("today: %v", got)
	}
	if got := todosShown(c.Day(d("2026-10-16"), today)); !sameShown(got, []shown{{"Call landlord", false, false}}) {
		t.Errorf("new day: %v", got)
	}
}

func TestDeletedTodoIsGone(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	todo := c.AddTodo(today, "Buy milk")
	c.DeleteTodo(todo.ID)
	if got := todosShown(c.Day(today, today)); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}
