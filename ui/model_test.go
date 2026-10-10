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
	for _, r := range m.listing().rows {
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
	// Buy milk went below the open tasks; the cursor kept its place.
	if it, ok := m.listing().current(); !ok || it.Title() != "Post letter" {
		t.Errorf("cursor is on %q, want Post letter", it.Title())
	}
}

func TestMoveCarriesATaskToThePickedDayAndEscPutsItBack(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "o", "Buy milk", "enter", "enter")

	press(m, "m", "l", "esc")
	if m.mode != modeNormal || m.selected != calendar.DateOf(now) || len(titles(m)) != 1 {
		t.Fatalf("after a cancelled move: mode %v, selected %v, rows %v", m.mode, m.selected, titles(m))
	}

	press(m, "enter", "m", "l", "l", "enter")
	if m.mode != modeNormal || m.status != "Moved “Buy milk” to Fri 9 Oct" {
		t.Errorf("mode %v, status %q", m.mode, m.status)
	}
	if got := titles(m); len(got) != 1 || got[0] != "Buy milk" {
		t.Errorf("the picked day shows %v", got)
	}
	press(m, "t")
	if got := titles(m); len(got) != 0 {
		t.Errorf("the task is still on its old day: %v", got)
	}
}

func TestRowsThatCannotBeMovedSayWhy(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "a", "Study wed", "enter", "o", "Buy milk", "enter", "enter")

	press(m, "m")
	if m.mode != modeNormal || m.status != "A recurring event can't be moved; edit its weekdays with e" {
		t.Errorf("m on a recurring event: mode %v, status %q", m.mode, m.status)
	}
	press(m, "j", " ")
	for _, key := range []string{"m", ">"} {
		press(m, key)
		if m.mode != modeNormal || m.status != "A done task can't be moved" {
			t.Errorf("%s on a done task: mode %v, status %q", key, m.mode, m.status)
		}
	}
	if got := titles(m); len(got) != 2 {
		t.Errorf("rows %v, want both still here", got)
	}
}

func TestDeletingAnOccurrenceAsksWhetherToSkipItOrDeleteAll(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "a", "Study wed", "enter", "enter", "d")
	if m.mode != modeDeleteChoice {
		t.Fatalf("mode %v, want the choice between this occurrence and all", m.mode)
	}
	press(m, "t")
	if got := titles(m); len(got) != 0 || m.status != "Skipped “Study” on Wed 7 Oct" {
		t.Errorf("after skipping: rows %v, status %q", got, m.status)
	}
	press(m, "j")
	if got := titles(m); len(got) != 1 {
		t.Fatalf("next week shows %v, want the next occurrence", got)
	}

	press(m, "enter", "d", "a")
	if m.mode != modeConfirm {
		t.Fatalf("mode %v, want a confirmation", m.mode)
	}
	press(m, "n")
	if got := titles(m); len(got) != 1 {
		t.Errorf("anything but y should keep the event: %v", got)
	}
	press(m, "d", "a", "y", "j")
	if got := titles(m); len(got) != 0 {
		t.Errorf("after deleting all, a later week shows %v", got)
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

func TestTaskIsAddedWithSubtasksThatAreTickedOnTheirOwn(t *testing.T) {
	m, path := newModel(t)
	resize(m, 120, 40)
	press(m, "o", "Pack s passport, charger")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Pack · 2 subtasks: passport, charger") {
		t.Errorf("no preview of how the line was read:\n%s", view)
	}
	press(m, "enter", "enter", "j", " ")

	view := ansi.Strip(m.View())
	for _, want := range []string{"[ ] Pack  1/2", "›     [x] passport", "      [ ] charger"} {
		if !strings.Contains(view, want) {
			t.Errorf("compact view lacks %q:\n%s", want, view)
		}
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), `"title": "passport"`) {
		t.Errorf("file does not have the subtask: %s", data)
	}

	press(m, "u")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "[ ] Pack  0/2") {
		t.Errorf("undo did not untick the subtask:\n%s", view)
	}
}

func TestEditingATaskLineKeepsTicksAndEditingASubtaskRenamesIt(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "o", "Pack s passport, charger", "enter", "enter", "j", " ")

	press(m, "k", "e")
	if got := m.input.String(); got != "Pack s passport, charger" {
		t.Fatalf("edit line = %q", got)
	}
	press(m, ", socks", "enter")
	if got := titles(m); strings.Join(got, "|") != "Pack|passport|charger|socks" {
		t.Fatalf("rows %v", got)
	}
	if view := ansi.Strip(m.View()); !strings.Contains(view, "[ ] Pack  1/3") || !strings.Contains(view, "[x] passport") {
		t.Errorf("passport lost its tick:\n%s", view)
	}

	press(m, "j", "e")
	if got := m.input.String(); got != "passport" {
		t.Fatalf("subtask edit line = %q", got)
	}
	press(m, ", visa", "enter")
	if m.mode != modeInput || !strings.Contains(m.status, "comma") {
		t.Errorf("a comma in a subtask should be refused: mode %v, status %q", m.mode, m.status)
	}
	press(m, "esc", "e", " and visa", "enter")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "[x] passport and visa") {
		t.Errorf("subtask was not renamed with its tick kept:\n%s", view)
	}
}

func TestSubtaskIsDeletedAloneAndMovesOnlyWithItsTask(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 120, 40)
	press(m, "o", "Pack s passport, charger", "enter", "enter", "j")

	for _, key := range []string{"m", ">"} {
		press(m, key)
		if m.mode != modeNormal || m.status != "A subtask moves with its task" {
			t.Errorf("%s on a subtask: mode %v, status %q", key, m.mode, m.status)
		}
	}
	press(m, "d")
	if got := titles(m); strings.Join(got, "|") != "Pack|charger" {
		t.Fatalf("after deleting a subtask: %v", got)
	}

	press(m, "k", ">", "l")
	if got := titles(m); strings.Join(got, "|") != "Pack|charger" {
		t.Errorf("the task should have moved with its subtask: %v", got)
	}
	press(m, "enter", "d")
	if got := titles(m); len(got) != 0 {
		t.Errorf("deleting the task should take its subtasks: %v", got)
	}
}

func TestFullViewShowsATaskWithItsCountButNotItsSubtasks(t *testing.T) {
	m, _ := newModel(t)
	resize(m, 160, 40)
	press(m, "o", "Pack s passport, charger", "enter", "o", "Buy milk", "enter")
	press(m, "enter", "j", " ", "j", "j") // tick passport, then on to Buy milk

	press(m, "v")
	full := ansi.Strip(m.View())
	if !strings.Contains(full, "[ ] Pack 1/2") || strings.Contains(full, "passport") {
		t.Errorf("full view should list Pack with its count and no subtasks:\n%s", full)
	}
	if !strings.Contains(full, "›[ ] Buy milk") {
		t.Errorf("the cursor should still be on Buy milk:\n%s", full)
	}
	press(m, "k")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Pack · 2 subtasks: passport, charger") {
		t.Errorf("bottom line should say what the task holds:\n%s", view)
	}

	press(m, "v")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "› [ ] Pack  1/2") {
		t.Fatalf("back in the compact view the cursor should be on Pack:\n%s", view)
	}
	press(m, " ")
	if view := ansi.Strip(m.View()); !strings.Contains(view, "[x] Pack  1/2") || strings.Contains(view, "passport") {
		t.Errorf("a done task should be one line without its subtasks:\n%s", view)
	}
}
