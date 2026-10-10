package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Matter121/grepdiem/calendar"
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
	cellWidth  = 6
	gridWidth  = 4 + 7*cellWidth // week number column + seven days
	gridHeight = 3 + 6*2         // title, gap, weekday names, six weeks of two lines
	boxExtra   = 4               // border and padding on both sides
	sideBySide = 80              // narrower than this, the panel goes under the grid
	weekColumn = 4               // the week number and its gap

	// Full view needs this width, which leaves a day about eight columns of title.
	fullViewWidth = 80
	minTitle      = 6 // an Event's time gives way before its title gets shorter than this
	bottomLines   = 2
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
	if m.full {
		body = m.fullView(today, m.width, bodyHeight)
	} else if m.width >= sideBySide {
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
			day := m.store.Calendar().Day(date, today)
			numbers += " " + m.dayNumber(day, today) + " "
			markers += "  " + markerRow(day.Rows)
		}
		b.WriteString("\n" + numbers + "\n" + markers)
	}
	return b.String()
}

// dayNumber draws a four-column day number: today in the accent colour, the
// selection inverted, other months faint, and underlined if it has open Tasks.
func (m *Model) dayNumber(day calendar.Day, today calendar.Date) string {
	style := m.dayStyle(day.Date, today)
	digits := strconv.Itoa(day.Date.Day)
	open := false
	for _, r := range day.Rows {
		open = open || (r.Kind == calendar.TaskRow && r.State != calendar.Done)
	}
	return style.Render(strings.Repeat(" ", 3-len(digits))) +
		style.Underline(open).Render(digits) + style.Render(" ")
}

func (m *Model) dayStyle(date, today calendar.Date) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case date == today:
		style = style.Bold(true).Foreground(accent)
	case date.Month != m.selected.Month:
		style = style.Faint(true)
	}
	if date == m.selected {
		style = style.Reverse(true)
		if m.mode == modeMove {
			style = style.Foreground(lipgloss.Color("3"))
		}
	}
	return style
}

// fullView is the month grid of the Full view: every day is outlined and
// lists its Events and Tasks under its number, in as many lines as the height
// allows. The selected day and today are outlined in their own colours.
func (m *Model) fullView(today calendar.Date, width, height int) string {
	sel := m.selected
	first := calendar.Date{Year: sel.Year, Month: sel.Month, Day: 1}
	start, last := first.Monday(), first.AddMonths(1).AddDays(-1)
	weeks := 0
	for !start.AddDays(7 * weeks).After(last) {
		weeks++
	}
	// A day's outline is shared with its neighbours, so a cell is the day's
	// left edge and its text, and the grid ends in one more edge.
	cell := (width - weekColumn - 1) / 7
	// Lines under each day's number; what doesn't divide goes to the first weeks.
	rows, spare := max((height-4)/weeks-1, 1), max(height-4, 0)%weeks

	names := strings.Repeat(" ", weekColumn)
	for _, name := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		names += fmt.Sprintf(" %-*s", cell-1, name)
	}
	lines := []string{
		lipgloss.PlaceHorizontal(width, lipgloss.Center, heading.Render(sel.Format("January 2006"))),
		"",
		faint.Render(names),
	}

	// Where days share an edge it takes the colour of the one that stands out
	// more: the selected day, then today, then any other.
	outlines := []lipgloss.Style{
		lipgloss.NewStyle().Foreground(dim),
		lipgloss.NewStyle().Foreground(accent),
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true),
	}
	level := func(week, col int) int {
		if week < 0 || week >= weeks || col < 0 || col >= 7 {
			return 0
		}
		switch start.AddDays(7*week + col) {
		case sel:
			return 2
		case today:
			return 1
		}
		return 0
	}
	// rule is the line above a week, carrying its day numbers, or the one
	// that closes the grid.
	rule := func(week int) string {
		var b strings.Builder
		for col := range 8 {
			// A corner joins only the edges of its own colour, so a coloured
			// outline turns there instead of reaching into the days around it.
			up := max(level(week-1, col-1), level(week-1, col))
			down := max(level(week, col-1), level(week, col))
			left := max(level(week-1, col-1), level(week, col-1))
			right := max(level(week-1, col), level(week, col))
			corner := max(up, down, left, right)
			b.WriteString(outlines[corner].Render(junction(
				week > 0 && up == corner, week < weeks && down == corner,
				col > 0 && left == corner, col < 7 && right == corner)))
			if col == 7 {
				break
			}
			label := ""
			if week < weeks {
				date := start.AddDays(7*week + col)
				label = " " + m.dayStyle(date, today).Render(strconv.Itoa(date.Day)) + " "
			}
			b.WriteString(label + outlines[right].Render(strings.Repeat("─", max(cell-1-lipgloss.Width(label), 0))))
		}
		return b.String()
	}

	for week := range weeks {
		monday := start.AddDays(7 * week)
		_, number := monday.ISOWeek()
		rows := rows
		if week < spare {
			rows++
		}
		block := make([]string, 1+rows)
		block[0] = faint.Render(fmt.Sprintf("%*d  ", weekColumn-2, number)) + rule(week)
		for i := 1; i <= rows; i++ {
			block[i] = strings.Repeat(" ", weekColumn)
		}
		for col := range 7 {
			side := outlines[max(level(week, col-1), level(week, col))].Render("│")
			for j, line := range m.dayBox(monday.AddDays(col), today, cell-1, rows) {
				pad := strings.Repeat(" ", max(cell-1-lipgloss.Width(line), 0))
				block[1+j] += side + line + pad
			}
		}
		for i := 1; i <= rows; i++ {
			block[i] += outlines[level(week, 6)].Render("│")
		}
		lines = append(lines, block...)
	}
	lines = append(lines, strings.Repeat(" ", weekColumn)+rule(weeks))
	return strings.Join(lines, "\n")
}

// junction is the box-drawing character joining outlines that leave a point
// in the given directions.
func junction(up, down, left, right bool) string {
	i := 0
	for bit, on := range []bool{up, down, left, right} {
		if on {
			i |= 1 << bit
		}
	}
	return string([]rune(" ╵╷│╴╯╮┤╶╰╭├─┴┬┼")[i])
}

// dayBox draws what is on one day of the Full view in rows lines, each at
// most width wide. A day with more than fits ends in "+2 more", and the
// day the cursor is in scrolls to keep the cursor's row in sight.
func (m *Model) dayBox(date, today calendar.Date, width, rows int) []string {
	list := m.store.Calendar().Day(date, today).WithoutSubtasks()
	otherMonth := date.Month != m.selected.Month
	cursor := -1
	if date == m.selected && m.focus == focusPanel {
		cursor = m.cursor
	}

	lines := make([]string, 0, rows)
	shown, top := len(list), 0
	if len(list) > rows {
		shown = rows - 1
		if shown == 0 && cursor >= 0 {
			shown = 1 // no room for both: the cursor's row matters more than the count
		}
		top = max(0, min(cursor-shown+1, len(list)-shown))
	}
	for i := top; i < top+shown; i++ {
		line := fullRow(list[i], otherMonth, width)
		if i == cursor {
			line = lipgloss.NewStyle().Foreground(accent).Bold(true).Render("›") + line
		}
		lines = append(lines, ansi.Truncate(line, width, "…"))
	}
	if hidden := len(list) - shown; hidden > 0 && len(lines) < rows {
		more := fmt.Sprintf("+%d more", hidden)
		if shown == 0 {
			more = fmt.Sprintf("%d items", hidden)
		}
		lines = append(lines, faint.Render(ansi.Truncate(more, width, "…")))
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return lines
}

// fullRow draws a Row as it appears in a day of the Full view. An Event's
// start time goes faintly before its title, unless in a day only width wide
// that would leave too little of the title to read.
func fullRow(r calendar.Row, otherMonth bool, width int) string {
	clock := ""
	if r.Kind == calendar.EventRow && r.Event.Start != "" {
		clock = calendar.ShortClock(r.Event.Start) + " "
		if width-2-len(clock) < min(lipgloss.Width(r.Title()), minTitle) {
			clock = ""
		}
	}
	count := ""
	if c := r.Count(); c != "" {
		count = faint.Render(" " + c)
	}
	switch {
	case r.Kind == calendar.EventRow && otherMonth:
		return faint.Render("● " + clock + r.Title())
	case r.Kind == calendar.EventRow:
		return marker(r.Event) + " " + faint.Render(clock) + r.Title()
	case r.State == calendar.Done:
		return faint.Render(r.Mark()+" ") + doneText.Render(r.Title()) + count
	case r.State == calendar.Overdue:
		return overdue.Render(r.Mark()+" "+r.Title()) + count
	case otherMonth:
		return faint.Render(r.Mark()+" "+r.Title()) + count
	}
	return r.Mark() + " " + r.Title() + count
}

// rowDetail says in full what a Row of the Full view had no room for.
func (m *Model) rowDetail(r calendar.Row) string {
	if r.Kind == calendar.EventRow {
		return calendar.Describe(r.Event.Spec(), m.selected)
	}
	detail := calendar.DescribeTask(r.Task.Spec())
	if note := r.Note(); note != "" {
		detail += " · " + note
	}
	return detail
}

// markerRow draws up to three Markers and a "+" if there are more Events.
func markerRow(rows []calendar.Row) string {
	var b strings.Builder
	width := 0
	for _, r := range rows {
		if r.Kind != calendar.EventRow {
			continue
		}
		if width == 3 {
			b.WriteString(faint.Render("+"))
			width++
			break
		}
		b.WriteString(marker(r.Event))
		width++
	}
	return b.String() + strings.Repeat(" ", 4-width)
}

func marker(e calendar.Event) string {
	return lipgloss.NewStyle().Foreground(markerColors[e.Color%calendar.Colors]).Render("●")
}

func (m *Model) panelView(today calendar.Date, width, height int) string {
	rows := m.store.Calendar().Day(m.selected, today).Rows
	_, week := m.selected.ISOWeek()
	sub := fmt.Sprintf("week %d", week)
	if m.selected == today {
		sub += " · today"
	}
	lines := []string{heading.Render(m.selected.Format("Mon 2 Jan 2006")), faint.Render(sub), ""}

	if len(rows) == 0 {
		lines = append(lines, faint.Render("Nothing here."), faint.Render("a adds an event, o adds a task."))
	}

	timeWidth := 0
	for _, r := range rows {
		timeWidth = max(timeWidth, lipgloss.Width(r.Time()))
	}

	cursorLine := 0
	for i, r := range rows {
		if i > 0 && r.Kind != calendar.EventRow && rows[i-1].Kind == calendar.EventRow {
			lines = append(lines, "")
		}
		line := "  "
		if m.focus == focusPanel && i == m.cursor {
			cursorLine = len(lines)
			line = lipgloss.NewStyle().Foreground(accent).Bold(true).Render("› ")
		}
		switch {
		case r.Kind == calendar.EventRow:
			line += marker(r.Event) + " "
			if timeWidth > 0 {
				line += fmt.Sprintf("%-*s ", timeWidth, r.Time())
			}
			line += r.Title()
			if r.Event.Recurring() {
				line += faint.Render(" ↻")
			}
		default:
			if r.Kind == calendar.SubtaskRow {
				line += "    "
			}
			if r.State == calendar.Done {
				line += faint.Render(r.Mark()+" ") + doneText.Render(r.Title())
			} else {
				line += r.Mark() + " " + r.Title()
			}
			if count := r.Count(); count != "" {
				line += faint.Render("  " + count)
			}
			if note := r.Note(); note != "" {
				line += overdue.Render("  " + note)
			}
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
			inputAddTask: "Task", inputEditTask: "Edit task", inputEditSubtask: "Edit subtask",
			inputGoto: "Go to",
		}[m.inputKind]
		prompt := heading.Render(label+" ›") + " "
		first = prompt + m.input.view(m.width-lipgloss.Width(prompt)-1)
		second = m.inputHint()
	case modeDeleteChoice:
		first = fmt.Sprintf("Delete “%s”:  t this occurrence   a all occurrences   esc cancel", m.target.Title())
	case modeConfirm:
		first = m.question + "  y yes   any other key no"
	case modeMove:
		first = fmt.Sprintf("Moving “%s”: pick a day", m.target.Title())
		second = "←↓↑→ day · H/L month · K/J year · t today · enter drop here · esc cancel"
	default:
		switch {
		case m.focus == focusPanel:
			second = "space tick · e edit · d delete · m move · esc back · ? help · j/k select · > next day · c colour"
			if rows := m.rows(); m.full && first == "" && m.cursor < len(rows) {
				first = m.rowDetail(rows[m.cursor])
			}
		case m.full:
			second = "a event · o task · enter open day · v compact view · ? help · q quit · ←↓↑→ day · t today · g go to · H/L month · K/J year · u undo"
		default:
			second = "a event · o task · enter open day · v full view · ? help · q quit · ←↓↑→ day · t today · g go to · H/L month · K/J year · u undo"
		}
	}
	return " " + ansi.Truncate(first, m.width-2, "…") + "\n " + faint.Render(ansi.Truncate(second, m.width-2, "…"))
}

// hintGap sets the format example apart from how the line was understood.
const hintGap = "     "

func (m *Model) inputHint() string {
	text := strings.TrimSpace(m.input.String())
	if m.status != "" {
		return m.status
	}
	switch m.inputKind {
	case inputAddEvent, inputEditEvent:
		format := "e.g.  Study tue thu 13-15 until w50   ·   Dentist 10:30"
		if text == "" {
			return format
		}
		return calendar.Describe(calendar.ParseEvent(text, m.selected), m.selected) + hintGap + format
	case inputAddTask, inputEditTask:
		format := "e.g.  Buy milk   ·   Pack s passport, charger, socks"
		if spec := calendar.ParseTask(text); len(spec.Subtasks) > 0 {
			return calendar.DescribeTask(spec) + hintGap + format
		}
		return format
	case inputGoto:
		if d, ok := calendar.ParseGoto(text, m.selected); ok {
			return d.Format("Mon 2 Jan 2006")
		}
		return "e.g.  14 oct   ·   2026-10-14   ·   w50   ·   dec   ·   2028"
	}
	return "enter save · esc cancel"
}

func (m *Model) helpView() string {
	keys := [][2]string{
		{"Anywhere", ""},
		{"← → / h l", "previous / next day"},
		{"shift+← → / H L", "previous / next month"},
		{"shift+↑ ↓ / K J", "previous / next year"},
		{"t", "jump to today"},
		{"g", "go to a date or week (14 oct, 2026-10-14, 14/10, w50, dec, 2028)"},
		{"a", "add an event on the selected day"},
		{"o", "add a task on the selected day"},
		{"u", "undo the last change"},
		{"v", "switch between the compact view and the full view"},
		{"q", "quit"},
		{"", ""},
		{"Calendar", ""},
		{"↑ ↓ / k j", "previous / next week"},
		{"enter / tab", "open the selected day"},
		{"", ""},
		{"Day", ""},
		{"↑ ↓ / k j", "select an event, task or subtask"},
		{"space", "tick or untick a task or subtask"},
		{"e", "edit (a task's line holds its subtasks; a subtask is edited alone)"},
		{"d", "delete (a recurring event asks: this occurrence or all)"},
		{"m", "move to another day (a subtask moves with its task)"},
		{">", "move a task to the next day"},
		{"c", "change an event's colour"},
		{"esc / tab", "back to the calendar"},
		{"", ""},
		{"Adding events", ""},
		{"Dentist 10:30", "one event at 10:30"},
		{"Lecture 9-11", "with an end time"},
		{"Study tue thu until w50", "every Tuesday and Thursday through week 50"},
		{"Work mon-sat 9-17", "every day from Monday to Saturday"},
		{"Gym mon until 13 dec", "also: until week 50, 2026-12-13, 261213, 13/12, dec (its first day), or no end"},
		{"Gym mon u 13 dec", "u is short for until"},
		{`"Apollo 13" fri 20`, "quotes keep the title exactly as typed"},
		{"", ""},
		{"Adding tasks", ""},
		{"Buy milk", "one task"},
		{"Pack s passport, charger", "with subtasks, separated by commas"},
		{"Pack subtasks passport", "s is short for subtasks"},
		{`"Season s finale" s snacks`, "quotes keep the title exactly as typed"},
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
