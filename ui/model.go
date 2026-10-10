// Package ui is the terminal interface.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Matter121/grepdiem/calendar"
	"github.com/Matter121/grepdiem/store"
)

type focus int

const (
	focusGrid focus = iota
	focusPanel
)

type mode int

const (
	modeNormal mode = iota
	modeInput
	modeDeleteChoice // "this occurrence" or "all"
	modeConfirm
	modeMove
	modeHelp
)

type inputKind int

const (
	inputAddEvent inputKind = iota
	inputAddTask
	inputEditEvent
	inputEditTask
	inputEditSubtask
	inputGoto
)

type Model struct {
	store *store.Store
	now   func() time.Time

	selected calendar.Date
	full     bool // Full view rather than Compact view
	focus    focus
	cursor   int // index into the selected day's Rows

	mode      mode
	input     lineInput
	inputKind inputKind
	target    calendar.Row  // what is being edited, deleted or moved
	moveFrom  calendar.Date // where to return if a move is cancelled
	question  string        // modeConfirm
	onYes     func()

	helpScroll int

	status string
	width  int
	height int
}

// New builds the UI on an open Store. now is the clock "today" comes from.
func New(st *store.Store, now func() time.Time) *Model {
	return &Model{store: st, now: now, selected: calendar.DateOf(now())}
}

type tickMsg struct{}

// tick redraws once a minute so "today" moves on at midnight.
func tick() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) Init() tea.Cmd { return tick() }

func (m *Model) today() calendar.Date { return calendar.DateOf(m.now()) }

// rows are what the cursor moves through: the selected day's Rows, without
// the Subtasks in the Full view, which has no room to list them.
func (m *Model) rows() []calendar.Row {
	day := m.store.Calendar().Day(m.selected, m.today())
	if m.full {
		return day.WithoutSubtasks()
	}
	return day.Rows
}

// change applies an edit and saves it. If that fails it says why in the
// status line and reports false.
func (m *Model) change(edit func(c *calendar.Calendar)) bool {
	if err := m.store.Change(edit); err != nil {
		m.status = sentence(err)
		return false
	}
	return true
}

// sentence turns an error into a status line.
func sentence(err error) string {
	s := err.Error()
	return strings.ToUpper(s[:1]) + s[1:]
}

func (m *Model) selectDay(d calendar.Date) {
	m.selected = d
	m.cursor = 0
}

// settle keeps the cursor and focus valid after the day's Rows change.
func (m *Model) settle() {
	n := len(m.rows())
	m.cursor = max(0, min(m.cursor, n-1))
	if n == 0 && m.mode != modeMove {
		m.focus = focusGrid
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.width < fullViewWidth {
			m.full = false
		}
	case tickMsg:
		return m, tick()
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		switch m.mode {
		case modeInput:
			m.updateInput(msg)
		case modeDeleteChoice:
			m.updateDeleteChoice(msg)
		case modeConfirm:
			m.mode = modeNormal
			if msg.String() == "y" {
				m.onYes()
			}
		case modeMove:
			m.updateMove(msg)
		case modeHelp:
			switch msg.String() {
			case "down", "j":
				m.helpScroll++
			case "up", "k":
				m.helpScroll = max(m.helpScroll-1, 0)
			default:
				m.mode = modeNormal
			}
		default:
			cmd = m.updateNormal(msg)
		}
		m.settle()
		return m, cmd
	}
	return m, nil
}

// navigate handles the keys that move between days. It reports whether key was one.
func (m *Model) navigate(key string, vertical bool) bool {
	switch key {
	case "left", "h":
		m.selectDay(m.selected.AddDays(-1))
	case "right", "l":
		m.selectDay(m.selected.AddDays(1))
	case "shift+left", "H":
		m.selectDay(m.selected.AddMonths(-1))
	case "shift+right", "L":
		m.selectDay(m.selected.AddMonths(1))
	case "shift+up", "K":
		m.selectDay(m.selected.AddMonths(-12))
	case "shift+down", "J":
		m.selectDay(m.selected.AddMonths(12))
	case "t":
		m.selectDay(m.today())
	case "up", "k":
		if !vertical {
			return false
		}
		m.selectDay(m.selected.AddDays(-7))
	case "down", "j":
		if !vertical {
			return false
		}
		m.selectDay(m.selected.AddDays(7))
	default:
		return false
	}
	return true
}

func (m *Model) updateNormal(msg tea.KeyMsg) tea.Cmd {
	m.status = ""
	key := msg.String()
	if m.navigate(key, m.focus == focusGrid) {
		return nil
	}

	switch key {
	case "q":
		return tea.Quit
	case "?":
		m.mode, m.helpScroll = modeHelp, 0
	case "a":
		m.openInput(inputAddEvent, "")
	case "o":
		m.openInput(inputAddTask, "")
	case "g":
		m.openInput(inputGoto, "")
	case "u":
		m.undoLast()
	case "v":
		m.toggleView()
	case "tab":
		if m.focus == focusGrid && len(m.rows()) > 0 {
			m.focus = focusPanel
		} else {
			m.focus = focusGrid
		}
	case "enter":
		if len(m.rows()) > 0 {
			m.focus = focusPanel
		}
	case "esc":
		m.focus = focusGrid
	}

	if m.focus != focusPanel {
		return nil
	}
	rows := m.rows()
	if m.cursor >= len(rows) {
		return nil
	}
	it := rows[m.cursor]
	switch key {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, len(rows)-1)
	case " ":
		switch it.Kind {
		case calendar.TaskRow:
			today := m.today()
			m.change(func(c *calendar.Calendar) { c.ToggleTask(it.Task.ID, today) })
		case calendar.SubtaskRow:
			m.change(func(c *calendar.Calendar) { c.ToggleSubtask(it.Task.ID, it.Sub) })
		}
	case "e":
		m.target = it
		switch it.Kind {
		case calendar.TaskRow:
			m.openInput(inputEditTask, calendar.TaskLine(it.Task.Spec()))
		case calendar.SubtaskRow:
			m.openInput(inputEditSubtask, it.Title())
		default:
			m.openInput(inputEditEvent, calendar.Line(it.Event.Spec()))
		}
	case "d":
		m.delete(it)
	case "c":
		if it.Kind == calendar.EventRow {
			m.change(func(c *calendar.Calendar) { c.CycleColor(it.Event.ID) })
		}
	case "m":
		m.startMove(it)
	case ">":
		if m.movable(it) && it.Kind == calendar.TaskRow {
			next := m.selected.AddDays(1)
			if m.change(func(c *calendar.Calendar) { c.MoveTask(it.Task.ID, next) }) {
				m.status = fmt.Sprintf("Moved “%s” to %s", it.Title(), next.Format("Mon 2 Jan"))
			}
		}
	}
	return nil
}

// toggleView switches between Compact view and Full view. The cursor stays
// on its Event or Task; from a Subtask it goes to the Task.
func (m *Model) toggleView() {
	if !m.full && m.width < fullViewWidth {
		m.status = fmt.Sprintf("Full view needs a terminal at least %d wide", fullViewWidth)
		return
	}
	var at calendar.Row
	if rows := m.rows(); m.cursor < len(rows) {
		at = rows[m.cursor]
	}
	m.full = !m.full
	for i, r := range m.rows() {
		if r.Kind != calendar.SubtaskRow && r.Event.ID == at.Event.ID && r.Task.ID == at.Task.ID {
			m.cursor = i
		}
	}
}

func (m *Model) undoLast() {
	undone, err := m.store.Undo()
	switch {
	case err != nil:
		m.status = sentence(err)
	case !undone:
		m.status = "Nothing to undo"
	default:
		m.status = "Undone (u again to redo)"
	}
}

func (m *Model) openInput(kind inputKind, text string) {
	m.mode, m.inputKind = modeInput, kind
	m.input.set(text)
}

func (m *Model) updateInput(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		m.status = ""
	case tea.KeyEnter:
		m.status = ""
		m.submit()
	default:
		m.input.handle(msg)
	}
}

func (m *Model) submit() {
	text := strings.TrimSpace(m.input.String())
	if text == "" {
		m.mode = modeNormal
		return
	}
	switch m.inputKind {
	case inputAddTask, inputEditTask:
		spec := calendar.ParseTask(text)
		if spec.Title == "" {
			m.status = "A task needs a title"
			return
		}
		day, id := m.selected, m.target.Task.ID
		m.change(func(c *calendar.Calendar) {
			if m.inputKind == inputAddTask {
				c.AddTask(day, spec.Title, spec.Subtasks...)
			} else {
				c.UpdateTask(id, spec.Title, spec.Subtasks...)
			}
		})
	case inputEditSubtask:
		if strings.Contains(text, ",") {
			m.status = "A subtask can't have a comma in it"
			return
		}
		it := m.target
		m.change(func(c *calendar.Calendar) { c.RenameSubtask(it.Task.ID, it.Sub, text) })
	case inputGoto:
		d, ok := calendar.ParseGoto(text, m.selected)
		if !ok {
			m.status = "Try “14 oct”, “2026-10-14” or “w50”"
			return
		}
		m.selectDay(d)
	case inputAddEvent:
		spec := calendar.ParseEvent(text, m.selected)
		if spec.Title == "" {
			m.status = "An event needs a title"
			return
		}
		day := m.selected
		var added calendar.Event
		if m.change(func(c *calendar.Calendar) { added = c.AddEvent(day, spec) }) {
			m.status = addedStatus(added, day)
		}
	case inputEditEvent:
		spec := calendar.ParseEvent(text, m.selected)
		if spec.Title == "" {
			m.status = "An event needs a title"
			return
		}
		old := m.target.Event
		update := func() { m.change(func(c *calendar.Calendar) { c.UpdateEvent(old.ID, spec) }) }
		if old.Recurring() && !spec.Recurring() {
			m.ask(fmt.Sprintf("Keep only %s and drop every other occurrence of “%s”?",
				old.Date.Format("Mon 2 Jan"), old.Title), update)
			return
		}
		update()
	}
	m.mode = modeNormal
}

// addedStatus says where a new Recurring Event first shows up, since that is
// often not the selected day.
func addedStatus(e calendar.Event, selected calendar.Date) string {
	if !e.Recurring() {
		return ""
	}
	first := e.First()
	switch {
	case first.IsZero():
		return fmt.Sprintf("Added “%s”, but it ends before its first occurrence", e.Title)
	case first != selected:
		return fmt.Sprintf("Added “%s”, first on %s", e.Title, first.Format("Mon 2 Jan"))
	}
	return ""
}

func (m *Model) ask(question string, onYes func()) {
	m.mode, m.question, m.onYes = modeConfirm, question, onYes
}

func (m *Model) delete(it calendar.Row) {
	switch {
	case it.Kind == calendar.TaskRow:
		m.change(func(c *calendar.Calendar) { c.DeleteTask(it.Task.ID) })
	case it.Kind == calendar.SubtaskRow:
		m.change(func(c *calendar.Calendar) { c.DeleteSubtask(it.Task.ID, it.Sub) })
	case it.Event.Recurring():
		m.target = it
		m.mode = modeDeleteChoice
	default:
		m.change(func(c *calendar.Calendar) { c.DeleteEvent(it.Event.ID) })
	}
}

func (m *Model) updateDeleteChoice(msg tea.KeyMsg) {
	m.mode = modeNormal
	e, day := m.target.Event, m.selected
	switch msg.String() {
	case "t":
		if m.change(func(c *calendar.Calendar) { c.Skip(e.ID, day) }) {
			m.status = fmt.Sprintf("Skipped “%s” on %s", e.Title, day.Format("Mon 2 Jan"))
		}
	case "a":
		m.ask(fmt.Sprintf("Delete every occurrence of “%s”?", e.Title), func() {
			m.change(func(c *calendar.Calendar) { c.DeleteEvent(e.ID) })
		})
	}
}

func (m *Model) movable(it calendar.Row) bool {
	switch {
	case it.Kind == calendar.SubtaskRow:
		m.status = "A subtask moves with its task"
	case it.Kind == calendar.EventRow && it.Event.Recurring():
		m.status = "A recurring event can't be moved; edit its weekdays with e"
	case it.State == calendar.Done:
		m.status = "A done task can't be moved"
	default:
		return true
	}
	return false
}

func (m *Model) startMove(it calendar.Row) {
	if !m.movable(it) {
		return
	}
	m.target, m.moveFrom = it, m.selected
	m.mode, m.focus = modeMove, focusGrid
}

func (m *Model) updateMove(msg tea.KeyMsg) {
	key := msg.String()
	if m.navigate(key, true) {
		return
	}
	switch key {
	case "enter":
		it, day := m.target, m.selected
		moved := m.change(func(c *calendar.Calendar) {
			if it.Kind == calendar.TaskRow {
				c.MoveTask(it.Task.ID, day)
			} else {
				c.MoveEvent(it.Event.ID, day)
			}
		})
		m.mode = modeNormal
		if moved {
			m.status = fmt.Sprintf("Moved “%s” to %s", it.Title(), day.Format("Mon 2 Jan"))
		}
	case "esc", "q":
		m.mode = modeNormal
		m.selectDay(m.moveFrom)
	}
}
