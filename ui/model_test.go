package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/Matter121/grepdiem/calendar"
	"github.com/Matter121/grepdiem/store"
)

var now = time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

func newModel(t *testing.T) (*Model, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return New(st, func() time.Time { return now }), path
}

// press sends keys: "enter" and "esc" by name, anything else as typed text.
func press(m *Model, keys ...string) {
	for _, key := range keys {
		switch key {
		case "enter":
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		case "esc":
			m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		default:
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		}
	}
}

func titles(m *Model) []string {
	var titles []string
	for _, r := range m.rows() {
		titles = append(titles, r.Title())
	}
	return titles
}

func TestAddedTaskIsSavedAndUndoTakesItBack(t *testing.T) {
	m, path := newModel(t)
	press(m, "o", "Buy milk", "enter")
	if got := titles(m); len(got) != 1 || got[0] != "Buy milk" {
		t.Fatalf("day shows %v, want Buy milk", got)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "Buy milk") {
		t.Errorf("file does not have the task: %s", data)
	}

	press(m, "u")
	if got := titles(m); len(got) != 0 {
		t.Errorf("after undo the day shows %v", got)
	}
	if data, _ := os.ReadFile(path); strings.Contains(string(data), "Buy milk") {
		t.Errorf("file still has the task: %s", data)
	}
}

func TestFailedChangeIsReportedInsteadOfSuccess(t *testing.T) {
	m, path := newModel(t)
	press(m, "o", "Buy milk", "enter")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	press(m, "enter", ">")
	if !strings.HasPrefix(m.status, "Could not reload") {
		t.Errorf("status = %q, want the reload failure", m.status)
	}
	if got := m.store.Calendar().Tasks[0].Date; got != calendar.DateOf(now) {
		t.Errorf("task moved to %v although the change failed", got)
	}
}

func TestUndoDoesNotOverwriteAChangeMadeElsewhere(t *testing.T) {
	m, path := newModel(t)
	press(m, "o", "Mine", "enter")
	other, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Change(func(c *calendar.Calendar) { c.AddTask(calendar.DateOf(now), "Theirs") }); err != nil {
		t.Fatal(err)
	}

	press(m, "u")
	if got := titles(m); len(got) != 2 {
		t.Errorf("day shows %v, want both tasks", got)
	}
	if !strings.Contains(m.status, "changed outside") {
		t.Errorf("status = %q, want it to say the file changed", m.status)
	}
}

func resize(m *Model, width, height int) {
	m.Update(tea.WindowSizeMsg{Width: width, Height: height})
}

func TestFullViewListsEveryDaysEventsAndTasks(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "a", "Dentist 10:30", "enter", "o", "Buy milk", "enter")
	press(m, "l", "a", "Lecture 9-11", "enter", "h")

	compact := ansi.Strip(m.View())
	if strings.Contains(compact, "Lecture") {
		t.Errorf("compact view shows another day's event:\n%s", compact)
	}

	press(m, "v")
	full := ansi.Strip(m.View())
	for _, want := range []string{"October 2026", "● 10:30 Dentist", "[ ] Buy milk", "● 9 Lecture"} {
		if !strings.Contains(full, want) {
			t.Errorf("full view lacks %q:\n%s", want, full)
		}
	}
	for i, line := range strings.Split(full, "\n") {
		if w := ansi.StringWidth(line); w > 120 {
			t.Errorf("line %d is %d wide", i, w)
		}
	}
	if n := strings.Count(full, "\n") + 1; n != 40 {
		t.Errorf("full view is %d lines, want 40", n)
	}

	press(m, "v")
	if m.full {
		t.Error("v did not switch back to the compact view")
	}
}

func TestFullViewDayWithMoreThanFits(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 100, 22) // two lines under each day's number
	for _, title := range []string{"One", "Two", "Three", "Four"} {
		press(m, "o", title, "enter")
	}
	press(m, "v")
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "[ ] One") || !strings.Contains(view, "+3 more") || strings.Contains(view, "Two") {
		t.Errorf("want the first task and +3 more:\n%s", view)
	}

	// Stepping in and down scrolls the day to the cursor.
	press(m, "enter", "j", "j", "j")
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "›[ ] Four") {
		t.Errorf("want the cursor on the fourth task:\n%s", view)
	}
}

func TestFullViewActsOnTheItemUnderTheCursor(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "a", "Dentist 10:30", "enter", "o", "Buy milk", "enter", "v", "enter")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Dentist · Wed 7 Oct · 10:30") {
		t.Errorf("bottom line lacks the event's details:\n%s", view)
	}

	press(m, "j", " ")
	if done := m.store.Calendar().Tasks[0].Done; done != calendar.DateOf(now) {
		t.Errorf("task done on %v, want today", done)
	}
	if view := ansi.Strip(m.View()); !strings.Contains(view, "[x] Buy milk") || strings.Contains(view, "· done") {
		t.Errorf("want the task ticked, and its tick box alone to say so:\n%s", view)
	}
}

func TestFullViewNeedsAWideTerminal(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 70, 30)
	press(m, "v")
	if m.full || !strings.Contains(m.status, "at least 80 wide") {
		t.Errorf("full = %v, status = %q", m.full, m.status)
	}

	resize(m, 100, 30)
	press(m, "v")
	if !m.full {
		t.Fatal("v did not open the full view")
	}
	resize(m, 70, 30)
	if m.full {
		t.Error("still in the full view after the terminal shrank")
	}
}

func TestShiftedKeysJumpMonthsAndYears(t *testing.T) {
	m, _ := newModel(t)
	steps := []struct {
		key  tea.KeyMsg
		want calendar.Date
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}, calendar.Date{Year: 2026, Month: 11, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyShiftLeft}, calendar.Date{Year: 2026, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("H")}, calendar.Date{Year: 2026, Month: 9, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyShiftRight}, calendar.Date{Year: 2026, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("J")}, calendar.Date{Year: 2027, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyShiftUp}, calendar.Date{Year: 2026, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")}, calendar.Date{Year: 2025, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyShiftDown}, calendar.Date{Year: 2026, Month: 10, Day: 7}},
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")}, calendar.Date{Year: 2026, Month: 10, Day: 7}},
	}
	for _, step := range steps {
		m.Update(step.key)
		if m.selected != step.want {
			t.Fatalf("after %q: selected %v, want %v", step.key.String(), m.selected, step.want)
		}
	}
}

func TestCompactViewCursorMarksTheRowAKeyActsOn(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "a", "Dentist 10:30", "enter", "o", "Buy milk", "enter", "o", "Post letter", "enter")

	press(m, "enter", "j")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "› [ ] Buy milk") {
		t.Fatalf("want the cursor on Buy milk:\n%s", view)
	}
	press(m, " ")
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "[x] Buy milk") || !strings.Contains(view, "[ ] Post letter") {
		t.Errorf("want only Buy milk ticked:\n%s", view)
	}
}

func TestOverdueAndDoneTasksShowInBothViews(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "h", "o", "Rent", "enter", "l", "o", "Plants", "enter")
	press(m, "enter", "j", " ", "esc")

	compact := ansi.Strip(m.View())
	for _, want := range []string{"[ ] Rent  overdue from Tue 6 Oct", "[x] Plants"} {
		if !strings.Contains(compact, want) {
			t.Errorf("compact view lacks %q:\n%s", want, compact)
		}
	}

	press(m, "v")
	full := ansi.Strip(m.View())
	for _, want := range []string{"[ ] Rent", "[x] Plants"} {
		if !strings.Contains(full, want) {
			t.Errorf("full view lacks %q:\n%s", want, full)
		}
	}
	press(m, "enter", "k")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Rent · overdue from Tue 6 Oct") {
		t.Errorf("bottom line lacks where the overdue task comes from:\n%s", view)
	}
}
