package calendar

import (
	"strconv"
	"strings"
	"time"
)

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

// parseDatePhrase reads a date written as one to three words: "w50",
// "week 50", "2026-12-13", "261213", "13/12", "13.12.2026", "13 dec",
// "dec 13", "13 dec 2026" or "dec". Anything left unsaid comes from the
// selected day, and a month on its own is the first of the next such month.
// As an end, a week or a date without a year that is already past means the
// next one, and a week means its Sunday rather than its Monday.
func parseDatePhrase(words []string, selected Date, asEnd bool) (Date, bool) {
	switch len(words) {
	case 1:
		if d, ok := parseWeek(words[0], selected, asEnd); ok {
			return d, true
		}
		if d, ok := parseTypedDate(words[0]); ok {
			return d, true
		}
		if d, ok := parseNumericDate(words[0], selected, asEnd); ok {
			return d, true
		}
		return parseMonth(words[0], selected, asEnd)
	case 2:
		if strings.EqualFold(words[0], "week") {
			return parseWeek("w"+words[1], selected, asEnd)
		}
		return parseDayMonth(words[0], words[1], selected, asEnd)
	case 3:
		if year, err := strconv.Atoi(words[2]); err == nil && len(words[2]) == 4 {
			return parseDayMonth(words[0], words[1], Date{Year: year}, false)
		}
	}
	return Date{}, false
}

// ParseGoto reads where to jump to: any date parseDatePhrase reads, or a year
// on its own ("2028"), which is the selected day in that year.
func ParseGoto(line string, selected Date) (Date, bool) {
	words := strings.Fields(line)
	if d, ok := parseDatePhrase(words, selected, false); ok {
		return d, true
	}
	if len(words) != 1 {
		return Date{}, false
	}
	if year, err := strconv.Atoi(words[0]); err == nil && len(words[0]) == 4 {
		return selected.AddMonths(12 * (year - selected.Year)), true
	}
	return Date{}, false
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

// parseMonth reads "dec" as the first of the next December, which is this
// year's unless that month is over. As an end, the first of the selected
// month is also past once the selected day is beyond it.
func parseMonth(word string, selected Date, asEnd bool) (Date, bool) {
	m, ok := monthNames[strings.ToLower(word)]
	if !ok {
		return Date{}, false
	}
	d := Date{selected.Year, m, 1}
	if m < selected.Month || (asEnd && d.Before(selected)) {
		d.Year++
	}
	return d, true
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

// parseNumericDate reads a day-first date such as "13/12", "13.12",
// "13/12/26" or "13.12.2026". Without a year it is the selected year; as an
// end, a date already past means next year.
func parseNumericDate(word string, selected Date, asEnd bool) (Date, bool) {
	sep := "/"
	if !strings.Contains(word, sep) {
		sep = "."
	}
	parts := strings.Split(word, sep)
	if len(parts) != 2 && len(parts) != 3 {
		return Date{}, false
	}
	var nums [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return Date{}, false
		}
		nums[i] = n
	}
	d := Date{selected.Year, time.Month(nums[1]), nums[0]}
	if len(parts) == 3 {
		switch len(parts[2]) {
		case 2:
			d.Year = 2000 + nums[2]
		case 4:
			d.Year = nums[2]
		default:
			return Date{}, false
		}
	}
	return settle(d, selected, asEnd && len(parts) == 2)
}

// settle moves a date already past into next year if rollover is set, and
// rejects one that doesn't exist, such as 31 Feb.
func settle(d, selected Date, rollover bool) (Date, bool) {
	if rollover && d.Before(selected) {
		d.Year++
	}
	if d.AddDays(0) != d {
		return Date{}, false
	}
	return d, true
}

// parseDayMonth reads "13 dec" or "dec 13", and with a year as the selected
// day "13 dec 2026". As an end, a date already past means next year.
func parseDayMonth(day, month string, selected Date, asEnd bool) (Date, bool) {
	if _, named := monthNames[strings.ToLower(day)]; named {
		day, month = month, day
	}
	n, err := strconv.Atoi(day)
	m, ok := monthNames[strings.ToLower(month)]
	if err != nil || !ok {
		return Date{}, false
	}
	return settle(Date{selected.Year, m, n}, selected, asEnd)
}
