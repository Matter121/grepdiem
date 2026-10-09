package calendar

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var weekdayNames = map[string]time.Weekday{
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
	"sun": time.Sunday, "sunday": time.Sunday,
}

var monthNames = map[string]time.Month{
	"jan": time.January, "january": time.January,
	"feb": time.February, "february": time.February,
	"mar": time.March, "march": time.March,
	"apr": time.April, "april": time.April,
	"may": time.May,
	"jun": time.June, "june": time.June,
	"jul": time.July, "july": time.July,
	"aug": time.August, "august": time.August,
	"sep": time.September, "sept": time.September, "september": time.September,
	"oct": time.October, "october": time.October,
	"nov": time.November, "november": time.November,
	"dec": time.December, "december": time.December,
}

// ParseEvent reads a Quick-add line for an Event on the selected day.
//
// Keywords (weekdays, a time, "until ...") only count at the end of the line:
// the line is read backwards and everything from the first non-keyword to the
// start is the title. The first word is always title unless the title is
// given in double quotes.
func ParseEvent(line string, selected Date) EventSpec {
	var spec EventSpec
	title, words, quoted := splitQuoted(line)
	keep := 1 // words that can never be keywords
	if quoted {
		keep = 0
	}

	end := len(words)
	for end > keep {
		if n := parseUntil(words[keep:end], selected, &spec); n > 0 {
			end -= n
		} else if days, ok := parseWeekdays(words[end-1]); ok {
			spec.Weekdays = append(spec.Weekdays, days...)
			end--
		} else if start, stop, ok := parseTimes(words[end-1]); ok && spec.Start == "" {
			spec.Start, spec.End = start, stop
			end--
		} else {
			break
		}
	}

	spec.Title = strings.TrimSpace(title + " " + strings.Join(words[:end], " "))
	if !spec.Until.IsZero() && len(spec.Weekdays) == 0 {
		spec.Weekdays = []time.Weekday{selected.Weekday()}
	}
	spec.Weekdays = sortWeekdays(spec.Weekdays)
	return spec
}

// splitQuoted returns a leading "quoted title", if any, and the other words.
func splitQuoted(line string) (title string, words []string, quoted bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, `"`) {
		if i := strings.Index(line[1:], `"`); i >= 0 {
			return line[1 : 1+i], strings.Fields(line[2+i:]), true
		}
	}
	return "", strings.Fields(line), false
}

// parseUntil recognises "until w50", "until week 50", "until 2026-12-13",
// "until 261213" or "until 13 dec" at the end of words and returns how many
// words it used.
func parseUntil(words []string, selected Date, spec *EventSpec) int {
	if !spec.Until.IsZero() {
		return 0
	}
	n := len(words)
	isUntil := func(i int) bool { return i >= 0 && strings.EqualFold(words[i], "until") }

	if isUntil(n - 2) {
		if d, ok := parseWeek(words[n-1], selected, true); ok {
			spec.Until = d
			return 2
		}
		if d, ok := parseTypedDate(words[n-1]); ok {
			spec.Until = d
			return 2
		}
	}
	if isUntil(n - 3) {
		if d, ok := parseWeekWords(words[n-2], words[n-1], selected, true); ok {
			spec.Until = d
			return 3
		}
		if d, ok := parseDayMonth(words[n-2], words[n-1], selected, true); ok {
			spec.Until = d
			return 3
		}
	}
	return 0
}

// parseWeek reads "w50". As an end it means the Sunday of the next such
// week; otherwise the Monday of that week in the selected year.
func parseWeek(word string, selected Date, asEnd bool) (Date, bool) {
	if len(word) < 2 || (word[0] != 'w' && word[0] != 'W') {
		return Date{}, false
	}
	week, err := strconv.Atoi(word[1:])
	if err != nil || week < 1 || week > 53 {
		return Date{}, false
	}
	year, _ := selected.ISOWeek()
	monday := func(year int) Date {
		return Date{year, time.January, 4}.Monday().AddDays((week - 1) * 7)
	}
	if !asEnd {
		return monday(year), true
	}
	sunday := monday(year).AddDays(6)
	if sunday.Before(selected) {
		sunday = monday(year + 1).AddDays(6)
	}
	return sunday, true
}

// parseTypedDate reads "2026-12-13" or its short form "261213".
func parseTypedDate(word string) (Date, bool) {
	if d, err := ParseDate(word); err == nil {
		return d, true
	}
	if len(word) != 6 {
		return Date{}, false
	}
	t, err := time.Parse("20060102", "20"+word)
	if err != nil {
		return Date{}, false
	}
	return DateOf(t), true
}

// parseWeekWords reads "week 50", the spelled-out form of "w50".
func parseWeekWords(word, number string, selected Date, asEnd bool) (Date, bool) {
	if !strings.EqualFold(word, "week") {
		return Date{}, false
	}
	return parseWeek("w"+number, selected, asEnd)
}

// parseDayMonth reads "13 dec". As an end, a date already past means next year.
func parseDayMonth(day, month string, selected Date, asEnd bool) (Date, bool) {
	n, err := strconv.Atoi(day)
	m, ok := monthNames[strings.ToLower(month)]
	if err != nil || !ok {
		return Date{}, false
	}
	d := Date{selected.Year, m, n}
	if asEnd && d.Before(selected) {
		d.Year++
	}
	if d.AddDays(0) != d { // 31 feb
		return Date{}, false
	}
	return d, true
}

func parseWeekdays(word string) ([]time.Weekday, bool) {
	var days []time.Weekday
	for _, part := range strings.Split(strings.ToLower(word), ",") {
		day, ok := weekdayNames[part]
		if !ok {
			return nil, false
		}
		days = append(days, day)
	}
	return days, true
}

// parseTimes reads "13", "13:30", "13-15" or "13:30-15:00".
func parseTimes(word string) (start, end Clock, ok bool) {
	from, to, ranged := strings.Cut(word, "-")
	if start, ok = parseClock(from); !ok {
		return "", "", false
	}
	if ranged {
		if end, ok = parseClock(to); !ok {
			return "", "", false
		}
	}
	return start, end, true
}

func parseClock(s string) (Clock, bool) {
	hh, mm, hasMinutes := strings.Cut(s, ":")
	if len(hh) == 0 || len(hh) > 2 || (hasMinutes && len(mm) != 2) {
		return "", false
	}
	hour, err := strconv.Atoi(hh)
	if err != nil || hour < 0 || hour > 23 {
		return "", false
	}
	minute := 0
	if hasMinutes {
		if minute, err = strconv.Atoi(mm); err != nil || minute < 0 || minute > 59 {
			return "", false
		}
	}
	return NewClock(hour, minute), true
}

// sortWeekdays orders Monday first and drops repeats.
func sortWeekdays(days []time.Weekday) []time.Weekday {
	if len(days) == 0 {
		return nil
	}
	sort.Slice(days, func(i, j int) bool { return (days[i]+6)%7 < (days[j]+6)%7 })
	out := days[:1]
	for _, day := range days[1:] {
		if day != out[len(out)-1] {
			out = append(out, day)
		}
	}
	return out
}

// Weekdays is a set of weekdays, written in JSON as ["tue","thu"].
type Weekdays []time.Weekday

func (w Weekdays) MarshalJSON() ([]byte, error) {
	names := make([]string, len(w))
	for i, day := range w {
		names[i] = strings.ToLower(day.String()[:3])
	}
	return json.Marshal(names)
}

func (w *Weekdays) UnmarshalJSON(data []byte) error {
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return err
	}
	*w = nil
	for _, name := range names {
		day, ok := weekdayNames[strings.ToLower(name)]
		if !ok {
			return fmt.Errorf("unknown weekday %q", name)
		}
		*w = append(*w, day)
	}
	return nil
}
