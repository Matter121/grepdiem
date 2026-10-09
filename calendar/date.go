package calendar

import (
	"fmt"
	"time"
)

// Date is a calendar day with no time of day or time zone.
// The zero Date means "no date".
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// DateOf returns the day t falls on in t's own location.
func DateOf(t time.Time) Date {
	y, m, day := t.Date()
	return Date{y, m, day}
}

// ParseDate reads a date written as 2006-01-02.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q", s)
	}
	return DateOf(t), nil
}

func (d Date) time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) IsZero() bool                { return d == Date{} }
func (d Date) Weekday() time.Weekday       { return d.time().Weekday() }
func (d Date) ISOWeek() (year, week int)   { return d.time().ISOWeek() }
func (d Date) AddDays(n int) Date          { return DateOf(d.time().AddDate(0, 0, n)) }
func (d Date) Format(layout string) string { return d.time().Format(layout) }
func (d Date) String() string              { return d.Format(time.DateOnly) }

func (d Date) Before(o Date) bool { return d.time().Before(o.time()) }
func (d Date) After(o Date) bool  { return d.time().After(o.time()) }

// AddMonths moves n months, clamping to the last day of a shorter month.
func (d Date) AddMonths(n int) Date {
	first := time.Date(d.Year, d.Month+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return Date{first.Year(), first.Month(), min(d.Day, last)}
}

// Monday returns the Monday of d's week.
func (d Date) Monday() Date {
	return d.AddDays(-(int(d.Weekday()) + 6) % 7)
}

func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, nil
	}
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*d = Date{}
		return nil
	}
	parsed, err := ParseDate(string(b))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// Clock is a time of day written as "15:04". The empty Clock means "no time".
type Clock string

// NewClock builds a Clock from an hour and minute.
func NewClock(hour, minute int) Clock {
	return Clock(fmt.Sprintf("%02d:%02d", hour, minute))
}
