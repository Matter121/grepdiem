package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Matter121/tuitodo/calendar"
)

// Colours are the terminal's own 16, so the app follows the user's theme.
var (
	accent       = lipgloss.Color("4")
	dim          = lipgloss.Color("8")
	warn         = lipgloss.Color("1")
	markerColors = [calendar.Colors]lipgloss.Color{"1", "2", "3", "4", "5", "6"}

	faint    = lipgloss.NewStyle().Faint(true)
	bold     = lipgloss.NewStyle().Bold(true)
	heading  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	overdue  = lipgloss.NewStyle().Foreground(warn)
	doneText = lipgloss.NewStyle().Faint(true).Strikethrough(true)
)

const (
	cellWidth   = 6
	gridWidth   = 4 + 7*cellWidth // week number column + seven days
	gridHeight  = 3 + 6*2         // title, gap, weekday names, six weeks of two lines
	boxExtra    = 4               // border and padding on both sides
	sideBySide  = 80              // narrower than this, the panel goes under the grid
	bottomLines = 2
)

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.width < gridWidth+boxExtra || m.height < gridHeight+2+bottomLines {
		return fmt.Sprintf("Terminal too small (%d×%d).\nNeeds at least %d×%d.",
			m.width, m.height, gridWidth+boxExtra, gridHeight+2+bottomLines)
	}
	if m.mode == modeHelp {
		return m.helpView()
	}

	today := m.today()
	bodyHeight := m.height - bottomLines
	gridFocused := m.focus == focusGrid

	var body string
	if m.width >= sideBySide {
		inner := bodyHeight - 2
		panelWidth := m.width - gridWidth - 2*boxExtra
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			box(m.gridView(today), gridWidth, inner, gridFocused),
			box(m.panelView(today, panelWidth, inner), panelWidth, inner, !gridFocused))
	} else {
		grid := lipgloss.PlaceHorizontal(m.width-boxExtra, lipgloss.Center, m.gridView(today))
		body = box(grid, m.width-boxExtra, gridHeight, gridFocused)
		if inner := bodyHeight - gridHeight - 4; inner >= 1 {
			panelWidth := m.width - boxExtra
			body = lipgloss.JoinVertical(lipgloss.Left, body,
				box(m.panelView(today, panelWidth, inner), panelWidth, inner, !gridFocused))
		}
	}
	return body + "\n" + m.bottomView()
}

// box frames content at an exact inner size, highlighting the focused pane.
func box(content string, width, height int, focused bool) string {
	color := dim
	if focused {
		color = accent
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(color).
		Padding(0, 1).Width(width + 2).Height(height).MaxHeight(height + 2).
		Render(content)
}

func (m *Model) gridView(today calendar.Date) string {
	sel := m.selected
	start := calendar.Date{Year: sel.Year, Month: sel.Month, Day: 1}.Monday()

	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(gridWidth, lipgloss.Center, heading.Render(sel.Format("January 2006"))))
	b.WriteString("\n\n")
	b.WriteString(faint.Render("     Mon   Tue   Wed   Thu   Fri   Sat   Sun"))

	for week := range 6 {
		monday := start.AddDays(7 * week)
		_, number := monday.ISOWeek()
		numbers := faint.Render(fmt.Sprintf("%2d  ", number))
		markers := "    "
		for i := range 7 {
			date := monday.AddDays(i)
			day := m.cal.Day(date, today)
			numbers += " " + m.dayNumber(day, today) + " "
			markers += "  " + markerRow(day.Events)
		}
		b.WriteString("\n" + numbers + "\n" + markers)
	}
	return b.String()
}

// dayNumber draws a four-column day number: today in the accent colour, the
// selection inverted, other months faint, and underlined if it has open Todos.
func (m *Model) dayNumber(day calendar.Day, today calendar.Date) string {
	style := lipgloss.NewStyle()
	switch {
	case day.Date == today:
		style = style.Bold(true).Foreground(accent)
	case day.Date.Month != m.selected.Month:
		style = style.Faint(true)
	}
	if day.Date == m.selected {
		style = style.Reverse(true)
		if m.mode == modeMove {
			style = style.Foreground(lipgloss.Color("3"))
		}
	}
	digits := strconv.Itoa(day.Date.Day)
	open := false
	for _, t := range day.Todos {
		open = open || t.Done.IsZero()
	}
	return style.Render(strings.Repeat(" ", 3-len(digits))) +
		style.Underline(open).Render(digits) + style.Render(" ")
}

// markerRow draws up to three Markers and a "+" if there are more Events.
func markerRow(events []calendar.Event) string {
	var b strings.Builder
	width := 0
	for i, e := range events {
		if i == 3 {
			b.WriteString(faint.Render("+"))
			width++
			break
		}
		b.WriteString(marker(e))
		width++
	}
	return b.String() + strings.Repeat(" ", 4-width)
}

func marker(e calendar.Event) string {
	return lipgloss.NewStyle().Foreground(markerColors[e.Color%calendar.Colors]).Render("●")
}

func (m *Model) panelView(today calendar.Date, width, height int) string {
	day := m.cal.Day(m.selected, today)
	_, week := m.selected.ISOWeek()
	sub := fmt.Sprintf("week %d", week)
	if m.selected == today {
		sub += " · today"
	}
	lines := []string{heading.Render(m.selected.Format("Mon 2 Jan 2006")), faint.Render(sub), ""}

	if len(day.Events) == 0 && len(day.Todos) == 0 {
		lines = append(lines, faint.Render("Nothing here."), faint.Render("a adds an event, o adds a todo."))
	}

	timeWidth := 0
	for _, e := range day.Events {
		timeWidth = max(timeWidth, lipgloss.Width(calendar.TimeRange(e.Start, e.End)))
	}

	cursorLine, index := 0, 0
	pointer := func() string {
		defer func() { index++ }()
		if m.focus == focusPanel && index == m.cursor {
			cursorLine = len(lines)
			return lipgloss.NewStyle().Foreground(accent).Bold(true).Render("› ")
		}
		return "  "
	}

	for _, e := range day.Events {
		line := pointer() + marker(e) + " "
		if timeWidth > 0 {
			line += fmt.Sprintf("%-*s ", timeWidth, calendar.TimeRange(e.Start, e.End))
		}
		line += e.Title
		if e.Recurring() {
			line += faint.Render(" ↻")
		}
		lines = append(lines, line)
	}
	if len(day.Events) > 0 && len(day.Todos) > 0 {
		lines = append(lines, "")
	}
	for _, t := range day.Todos {
		line := pointer()
		switch {
		case !t.Done.IsZero():
			line += faint.Render("[x] ") + doneText.Render(t.Title)
		case t.Overdue:
			line += "[ ] " + t.Title + overdue.Render("  from "+t.Date.Format("Mon 2 Jan"))
		default:
			line += "[ ] " + t.Title
		}
		lines = append(lines, line)
	}

	// Scroll so the cursor's line stays visible.
	if len(lines) > height {
		top := max(0, min(cursorLine-height/2, len(lines)-height))
		lines = lines[top : top+height]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "…")
	}
	return strings.Join(lines, "\n")
}

// bottomView is two lines: a message or text field, then key hints.
func (m *Model) bottomView() string {
	first, second := m.status, ""
	switch m.mode {
	case modeInput:
		label := map[inputKind]string{
			inputAddEvent: "Event", inputEditEvent: "Edit event",
			inputAddTodo: "Todo", inputEditTodo: "Edit todo", inputGoto: "Go to",
		}[m.inputKind]
		prompt := heading.Render(label+" ›") + " "
		first = prompt + m.input.view(m.width-lipgloss.Width(prompt)-1)
		second = m.inputHint()
	case modeDeleteChoice:
		first = fmt.Sprintf("Delete “%s”:  t this occurrence   a all occurrences   esc cancel", m.target.title())
	case modeConfirm:
		first = m.question + "  y yes   any other key no"
	case modeMove:
		first = fmt.Sprintf("Moving “%s”: pick a day", m.target.title())
		second = "←↓↑→ day · [ ] month · t today · enter drop here · esc cancel"
	default:
		if m.focus == focusPanel {
			second = "space tick · e edit · d delete · m move · esc back · ? help · j/k select · > next day · c colour"
		} else {
			second = "a event · o todo · enter open day · ? help · q quit · ←↓↑→ day · t today · g go to · [ ] month · u undo"
		}
	}
	return " " + ansi.Truncate(first, m.width-2, "…") + "\n " + faint.Render(ansi.Truncate(second, m.width-2, "…"))
}

func (m *Model) inputHint() string {
	text := strings.TrimSpace(m.input.String())
	if m.status != "" {
		return m.status
	}
	switch m.inputKind {
	case inputAddEvent, inputEditEvent:
		if text == "" {
			return "e.g.  Study tue thu 13-15 until w50   ·   Dentist 10:30"
		}
		return calendar.Describe(calendar.ParseEvent(text, m.selected), m.selected)
	case inputGoto:
		if d, ok := calendar.ParseGoto(text, m.selected); ok {
			return d.Format("Mon 2 Jan 2006")
		}
		return "e.g.  14 oct   ·   2026-10-14   ·   w50"
	}
	return "enter save · esc cancel"
}

func (m *Model) helpView() string {
	keys := [][2]string{
		{"Anywhere", ""},
		{"← → / h l", "previous / next day"},
		{"[ ]", "previous / next month"},
		{"t", "jump to today"},
		{"g", "go to a date or week (14 oct, 2026-10-14, w50)"},
		{"a", "add an event on the selected day"},
		{"o", "add a todo on the selected day"},
		{"u", "undo the last change"},
		{"q", "quit"},
		{"", ""},
		{"Calendar", ""},
		{"↑ ↓ / k j", "previous / next week"},
		{"enter / tab", "open the selected day"},
		{"", ""},
		{"Day", ""},
		{"↑ ↓ / k j", "select an event or todo"},
		{"space", "tick or untick a todo"},
		{"e", "edit"},
		{"d", "delete (a recurring event asks: this occurrence or all)"},
		{"m", "move to another day"},
		{">", "move a todo to the next day"},
		{"c", "change an event's colour"},
		{"esc / tab", "back to the calendar"},
		{"", ""},
		{"Adding events", ""},
		{"Dentist 10:30", "one event at 10:30"},
		{"Lecture 9-11", "with an end time"},
		{"Study tue thu until w50", "every Tuesday and Thursday through week 50"},
		{"Gym mon until 13 dec", "also: until week 50, until 2026-12-13, or no end at all"},
		{`"Apollo 13" fri 20`, "quotes keep the title exactly as typed"},
	}
	var b strings.Builder
	for _, k := range keys {
		if k[1] == "" {
			b.WriteString(heading.Render(k[0]) + "\n")
			continue
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", bold.Render(fmt.Sprintf("%-24s", k[0])), k[1]))
	}
	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	footer := "Press any key to close."
	if room := m.height - 2; len(lines) > room {
		m.helpScroll = min(m.helpScroll, len(lines)-room)
		lines = lines[m.helpScroll : m.helpScroll+room]
		footer = "↑ ↓ scroll · any other key closes"
	}
	lines = append(lines, "", faint.Render(footer))
	for i, line := range lines {
		lines[i] = ansi.Truncate(" "+line, m.width, "…")
	}
	return strings.Join(lines, "\n")
}
