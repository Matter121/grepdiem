package store

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Matter121/grepdiem/calendar"
)

var day = calendar.Date{Year: 2026, Month: 10, Day: 7}

func open(t *testing.T, path string) *Store {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func addTask(t *testing.T, s *Store, title string) {
	t.Helper()
	if err := s.Change(func(c *calendar.Calendar) { c.AddTask(day, title) }); err != nil {
		t.Fatal(err)
	}
}

func tasks(c *calendar.Calendar) []string {
	var titles []string
	for _, task := range c.Tasks {
		titles = append(titles, task.Title)
	}
	return titles
}

// expect checks both what s holds and what is in its file.
func expect(t *testing.T, s *Store, want ...string) {
	t.Helper()
	if got := tasks(s.Calendar()); !slices.Equal(got, want) {
		t.Errorf("calendar has %v, want %v", got, want)
	}
	if got := tasks(open(t, s.path).Calendar()); !slices.Equal(got, want) {
		t.Errorf("file has %v, want %v", got, want)
	}
}

func TestMissingFileIsAnEmptyCalendar(t *testing.T) {
	s := open(t, filepath.Join(t.TempDir(), "sub", "data.json"))
	expect(t, s)
}

func TestChangeIsSaved(t *testing.T) {
	s := open(t, filepath.Join(t.TempDir(), "sub", "data.json"))
	addTask(t, s, "a")
	addTask(t, s, "b")
	expect(t, s, "a", "b")
}

func TestUndoTakesBackTheLastChangeAndASecondUndoRedoesIt(t *testing.T) {
	s := open(t, filepath.Join(t.TempDir(), "data.json"))
	if undone, err := s.Undo(); undone || err != nil {
		t.Fatalf("Undo with nothing to undo = %v, %v", undone, err)
	}
	addTask(t, s, "a")
	addTask(t, s, "b")

	if undone, err := s.Undo(); !undone || err != nil {
		t.Fatalf("Undo = %v, %v", undone, err)
	}
	expect(t, s, "a")

	if undone, err := s.Undo(); !undone || err != nil {
		t.Fatalf("second Undo = %v, %v", undone, err)
	}
	expect(t, s, "a", "b")
}

func TestChangeIsAppliedOnTopOfAChangeMadeElsewhere(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s, other := open(t, path), open(t, path)
	addTask(t, s, "mine")
	addTask(t, other, "theirs")
	addTask(t, s, "mine too")
	expect(t, s, "mine", "theirs", "mine too")

	// Undo takes back only the one change, not theirs.
	if undone, err := s.Undo(); !undone || err != nil {
		t.Fatalf("Undo = %v, %v", undone, err)
	}
	expect(t, s, "mine", "theirs")
}

func TestUndoKeepsAChangeMadeElsewhere(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s := open(t, path)
	addTask(t, s, "mine")
	addTask(t, open(t, path), "theirs")

	if undone, err := s.Undo(); undone || !errors.Is(err, ErrChangedElsewhere) {
		t.Fatalf("Undo = %v, %v, want ErrChangedElsewhere", undone, err)
	}
	expect(t, s, "mine", "theirs")

	if undone, err := s.Undo(); undone || err != nil {
		t.Fatalf("Undo after that = %v, %v, want nothing to undo", undone, err)
	}
	expect(t, s, "mine", "theirs")
}

func TestChangeIsNotAppliedWhenTheFileCannotBeReloaded(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	s := open(t, path)
	addTask(t, s, "a")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	applied := false
	err := s.Change(func(c *calendar.Calendar) { applied = true })
	if err == nil || applied {
		t.Errorf("Change = %v, applied = %v; want an error and no edit", err, applied)
	}
	if data, _ := os.ReadFile(path); string(data) != "not json" {
		t.Errorf("file was overwritten with %q", data)
	}
}

func TestChangeReportsAFailedSave(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub")
	s := open(t, filepath.Join(dir, "data.json"))
	// A file where the data directory should be.
	if err := os.WriteFile(dir, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	err := s.Change(func(c *calendar.Calendar) { c.AddTask(day, "a") })
	if err == nil {
		t.Fatal("Change = nil, want an error")
	}
	// The edit is kept and written by the next Change that succeeds.
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	addTask(t, s, "b")
	expect(t, s, "a", "b")
}
