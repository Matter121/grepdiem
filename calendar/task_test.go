package calendar

import "testing"

type shown struct {
	title string
	state TaskState
}

func tasksShown(day Day) []shown {
	var out []shown
	for _, r := range day.Rows {
		if r.Kind == TaskRow {
			out = append(out, shown{r.Title(), r.State})
		}
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

func TestTaskShowsOnItsDate(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	c.AddTask(today, "Buy milk")
	c.AddTask(d("2026-10-12"), "Hand in essay")

	if got := tasksShown(c.Day(today, today)); !sameShown(got, []shown{{"Buy milk", Open}}) {
		t.Errorf("today: %v", got)
	}
	if got := tasksShown(c.Day(d("2026-10-12"), today)); !sameShown(got, []shown{{"Hand in essay", Open}}) {
		t.Errorf("future day: %v", got)
	}
}

func TestUnfinishedTaskRollsOverToTodayAsOverdue(t *testing.T) {
	c := New()
	c.AddTask(d("2026-10-06"), "Call landlord")
	c.AddTask(d("2026-10-09"), "Buy milk")
	today := d("2026-10-09")

	want := []shown{{"Call landlord", Overdue}, {"Buy milk", Open}}
	if got := tasksShown(c.Day(today, today)); !sameShown(got, want) {
		t.Errorf("today: got %v, want %v", got, want)
	}
	if got := tasksShown(c.Day(d("2026-10-06"), today)); len(got) != 0 {
		t.Errorf("the day it was attached to should no longer show it: %v", got)
	}
}

func TestTickedOverdueTaskIsDoneOnTheDayItWasTicked(t *testing.T) {
	c := New()
	task := c.AddTask(d("2026-10-06"), "Call landlord")
	today := d("2026-10-09")
	c.toggleTask(task.ID, today)

	if got := tasksShown(c.Day(today, today)); !sameShown(got, []shown{{"Call landlord", Done}}) {
		t.Errorf("day ticked: %v", got)
	}
	// Still visible there the next day, and not Overdue again.
	tomorrow := d("2026-10-10")
	if got := tasksShown(c.Day(today, tomorrow)); !sameShown(got, []shown{{"Call landlord", Done}}) {
		t.Errorf("day ticked, seen tomorrow: %v", got)
	}
	if got := tasksShown(c.Day(tomorrow, tomorrow)); len(got) != 0 {
		t.Errorf("tomorrow: %v", got)
	}
}

func TestUntickingMakesATaskOpenAgain(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	task := c.AddTask(today, "Buy milk")
	c.toggleTask(task.ID, today)
	c.toggleTask(task.ID, today)
	if got := tasksShown(c.Day(today, today)); !sameShown(got, []shown{{"Buy milk", Open}}) {
		t.Errorf("got %v", got)
	}
}

func TestOpenTasksComeBeforeDoneOnes(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	first := c.AddTask(today, "First")
	c.AddTask(today, "Second")
	c.toggleTask(first.ID, today)

	want := []shown{{"Second", Open}, {"First", Done}}
	if got := tasksShown(c.Day(today, today)); !sameShown(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMovedOverdueTaskIsAnOrdinaryTaskOnItsNewDay(t *testing.T) {
	c := New()
	c.AddTask(d("2026-10-06"), "Call landlord")
	today := d("2026-10-09")
	if err := c.Move(row(t, c, today, "Call landlord"), d("2026-10-16")); err != nil {
		t.Fatal(err)
	}

	if got := tasksShown(c.Day(today, today)); len(got) != 0 {
		t.Errorf("today: %v", got)
	}
	if got := tasksShown(c.Day(d("2026-10-16"), today)); !sameShown(got, []shown{{"Call landlord", Open}}) {
		t.Errorf("new day: %v", got)
	}
}

func TestDeletedTaskIsGone(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	c.AddTask(today, "Buy milk")
	c.Delete(row(t, c, today, "Buy milk"))
	if got := tasksShown(c.Day(today, today)); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}
