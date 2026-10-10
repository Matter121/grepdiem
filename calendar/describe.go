package calendar

import (
	"reflect"
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
		times := ShortClock(spec.Start)
		if spec.End != "" {
			times += "-" + ShortClock(spec.End)
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

// ShortClock writes a Clock as briefly as it can be typed: "9", "13:30".
func ShortClock(c Clock) string {
	s := strings.TrimPrefix(string(c), "0")
	return strings.TrimSuffix(s, ":00")
}
