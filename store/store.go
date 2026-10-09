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

	"github.com/Matter121/TUICalendar/calendar"
)

// DefaultPath is $TUITODO_DATA, or tuitodo/data.json in the XDG data directory.
func DefaultPath() (string, error) {
	if p := os.Getenv("TUITODO_DATA"); p != "" {
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
	return filepath.Join(dir, "tuitodo", "data.json"), nil
}

// Store is the data file and what it looked like when last read or written.
type Store struct {
	path    string
	modTime time.Time
	size    int64
}

func New(path string) *Store { return &Store{path: path} }

func (s *Store) Path() string { return s.path }

// Load reads the Calendar. A missing file is an empty Calendar.
func (s *Store) Load() (*calendar.Calendar, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		s.modTime, s.size = time.Time{}, 0
		return calendar.New(), nil
	}
	if err != nil {
		return nil, err
	}
	c := calendar.New()
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("%s: %w", s.path, err)
	}
	s.remember()
	return c, nil
}

// Changed reports whether the file was modified by someone else since the
// last Load or Save.
func (s *Store) Changed() bool {
	info, err := os.Stat(s.path)
	if err != nil {
		return !s.modTime.IsZero()
	}
	return !info.ModTime().Equal(s.modTime) || info.Size() != s.size
}

// Save writes the Calendar atomically: a temp file renamed over the old one.
func (s *Store) Save(c *calendar.Calendar) error {
	data, err := json.MarshalIndent(c, "", "  ")
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
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return err
	}
	s.remember()
	return nil
}

func (s *Store) remember() {
	if info, err := os.Stat(s.path); err == nil {
		s.modTime, s.size = info.ModTime(), info.Size()
	}
}
