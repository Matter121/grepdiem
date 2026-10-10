// Package store keeps the Calendar in a single JSON file.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/Matter121/grepdiem/calendar"
)

// DefaultPath is $GREPDIEM_DATA, or grepdiem/data.json in the XDG data directory.
func DefaultPath() (string, error) {
	if p := os.Getenv("GREPDIEM_DATA"); p != "" {
		return p, nil
	}
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "grepdiem", "data.json"), nil
}

// ErrChangedElsewhere is returned by Undo when the data file was modified by
// someone else, so that undoing would have thrown their change away.
var ErrChangedElsewhere = errors.New("the data file changed outside grepdiem, so there is nothing to undo")

// Store is the Calendar and the data file it is kept in. Every change goes
// through Change or Undo, which keep the two in step.
type Store struct {
	path string
	cal  *calendar.Calendar
	undo *calendar.Calendar // the Calendar before the last change

	// What the file looked like when last read or written.
	modTime time.Time
	size    int64
}

// Open reads the Calendar at path. A missing file is an empty Calendar.
func Open(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Calendar returns the Calendar as of the last Open, Change or Undo. It is
// for reading: edits made outside Change are not saved.
func (s *Store) Calendar() *calendar.Calendar { return s.cal }

// Change applies an edit and saves it. If the file changed on disk since it
// was read, the edit is applied on top of what is there now. If that reload
// fails the edit is not applied; if the save fails the edit stays in the
// Calendar and is written by the next Change that succeeds.
func (s *Store) Change(edit func(c *calendar.Calendar)) error {
	if s.changed() {
		if err := s.load(); err != nil {
			return fmt.Errorf("could not reload: %w", err)
		}
	}
	s.undo = s.cal.Clone()
	edit(s.cal)
	return s.save()
}

// Undo takes back the last change and saves; a second Undo redoes it. It
// reports false if there is nothing to undo. If the file changed on disk
// since the last change, Undo keeps what is there now and returns
// ErrChangedElsewhere.
func (s *Store) Undo() (bool, error) {
	if s.changed() {
		hadUndo := s.undo != nil
		if err := s.load(); err != nil {
			return false, fmt.Errorf("could not reload: %w", err)
		}
		if hadUndo {
			return false, ErrChangedElsewhere
		}
	}
	if s.undo == nil {
		return false, nil
	}
	s.cal, s.undo = s.undo, s.cal
	return true, s.save()
}

// load reads the file and forgets the last change, which no longer applies.
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		s.cal, s.undo = calendar.New(), nil
		s.modTime, s.size = time.Time{}, 0
		return nil
	}
	if err != nil {
		return err
	}
	c := calendar.New()
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("%s: %w", s.path, err)
	}
	s.cal, s.undo = c, nil
	s.remember()
	return nil
}

// changed reports whether the file was modified by someone else since it was
// last read or written.
func (s *Store) changed() bool {
	info, err := os.Stat(s.path)
	if err != nil {
		return !s.modTime.IsZero()
	}
	return !info.ModTime().Equal(s.modTime) || info.Size() != s.size
}

// save writes the Calendar atomically: a temp file renamed over the old one.
func (s *Store) save() error {
	if err := s.write(); err != nil {
		return fmt.Errorf("could not save: %w", err)
	}
	s.remember()
	return nil
}

func (s *Store) write() error {
	data, err := json.MarshalIndent(s.cal, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".data-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

func (s *Store) remember() {
	if info, err := os.Stat(s.path); err == nil {
		s.modTime, s.size = info.ModTime(), info.Size()
	}
}
