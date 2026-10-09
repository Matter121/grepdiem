package calendar

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Describe says in words how a Quick-add line for the selected day was understood.
func Describe(spec EventSpec, selected Date) string {
	parts := []string{spec.Title}
	if spec.Recurring() {
		var days []string
		for _, day := range spec.Weekdays {
			days = append(days, day.String()[:3])
		}
		parts = append(parts, strings.Join(days, ", "))
	} else {
		parts = append(parts, selected.Format("Mon 2 Jan"))
	}
	if spec.Start != "" {
		parts = append(parts, TimeRange(spec.Start, spec.End))
	}
	if spec.Recurring() {
		switch {
		case spec.Until.IsZero():
			parts = append(parts, "no end")
		case spec.Until.Year != selected.Year:
			parts = append(parts, "until "+spec.Until.Format("Mon 2 Jan 2006"))
		default:
			parts = append(parts, "until "+spec.Until.Format("Mon 2 Jan"))
		}
	}
	return strings.Join(parts, " · ")
}

// TimeRange formats "13:00" or "13:00–15:00".
func TimeRange(start, end Clock) string {
	if end == "" {
		return string(start)
	}
	return string(start) + "–" + string(end)
}

// Line returns the Quick-add line that recreates spec.
func Line(spec EventSpec) string {
	var keywords []string
	for _, day := range spec.Weekdays {
		keywords = append(keywords, strings.ToLower(day.String()[:3]))
	}
	if spec.Start != "" {
		times := shortClock(spec.Start)
		if spec.End != "" {
			times += "-" + shortClock(spec.End)
		}
		keywords = append(keywords, times)
	}
	if !spec.Until.IsZero() {
		keywords = append(keywords, "until", spec.Until.String())
	}
	tail := strings.Join(keywords, " ")

	line := strings.TrimSpace(spec.Title + " " + tail)
	// Quote a title that would otherwise be read as keywords. The parse date
	// doesn't matter: Line never writes anything relative to it.
	if !reflect.DeepEqual(ParseEvent(line, Date{2000, time.January, 1}), spec) {
		line = strings.TrimSpace(`"` + spec.Title + `" ` + tail)
	}
	return line
}

func shortClock(c Clock) string {
	s := strings.TrimPrefix(string(c), "0")
	return strings.TrimSuffix(s, ":00")
}

// ParseGoto reads where to jump to: "14 oct", "oct 14", "14 oct 2027",
// "2026-10-14", "261014", "14/10", "14.10.2026", "w50" or "week 50". A month
// on its own ("dec") is the first of the next such month, and a year on its
// own ("2028") is the selected day in that year.
func ParseGoto(line string, selected Date) (Date, bool) {
	words := strings.Fields(line)
	switch len(words) {
	case 1:
		if d, ok := parseWeek(words[0], selected, false); ok {
			return d, true
		}
		if d, ok := parseTypedDate(words[0]); ok {
			return d, true
		}
		if d, ok := parseNumericDate(words[0], selected, false); ok {
			return d, true
		}
		if m, ok := monthNames[strings.ToLower(words[0])]; ok {
			d := Date{selected.Year, m, 1}
			if m < selected.Month {
				d.Year++
			}
			return d, true
		}
		if year, err := strconv.Atoi(words[0]); err == nil && len(words[0]) == 4 {
			return selected.AddMonths(12 * (year - selected.Year)), true
		}
	case 2:
		if d, ok := parseWeekWords(words[0], words[1], selected, false); ok {
			return d, true
		}
		return parseDayMonth(words[0], words[1], selected, false)
	case 3:
		return parseDayMonthYear(words[0], words[1], words[2])
	}
	return Date{}, false
}
