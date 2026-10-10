// Package ui is the terminal interface.
package ui

import (
	"errors"
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
	cursor   int // where in the selected day's listing the cursor is

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
	l := m.listing()
	m.cursor = l.cursor
	if len(l.rows) == 0 && m.mode != modeMove {
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
		if m.focus == focusGrid && len(m.listing().rows) > 0 {
			m.focus = focusPanel
		} else {
			m.focus = focusGrid
		}
	case "enter":
		if len(m.listing().rows) > 0 {
			m.focus = focusPanel
		}
	case "esc":
		m.focus = focusGrid
	}

	l := m.listing()
	it, ok := l.current()
	if !ok {
		return nil
	}
	switch key {
	case "up", "k":
		m.cursor = l.step(-1)
	case "down", "j":
		m.cursor = l.step(1)
	case " ":
		if it.Kind != calendar.EventRow {
			today := m.today()
			m.change(func(c *calendar.Calendar) { c.Tick(it, today) })
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
			m.change(func(c *calendar.Calendar) { c.CycleColor(it) })
		}
	case "m":
		m.startMove(it)
	case ">":
		if it.Kind != calendar.EventRow {
			m.move(it, m.selected.AddDays(1))
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
	if l := m.listing(); len(l.rows) > 0 {
		at = l.rows[l.cursor]
	}
	m.full = !m.full
	m.cursor = m.listing().find(at)
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
	if it.Occurrence() {
		m.target, m.mode = it, modeDeleteChoice
		return
	}
	m.change(func(c *calendar.Calendar) { c.Delete(it) })
}

func (m *Model) updateDeleteChoice(msg tea.KeyMsg) {
	m.mode = modeNormal
	it := m.target
	switch msg.String() {
	case "t":
		if m.change(func(c *calendar.Calendar) { c.Skip(it) }) {
			m.status = fmt.Sprintf("Skipped “%s” on %s", it.Title(), it.Date.Format("Mon 2 Jan"))
		}
	case "a":
		m.ask(fmt.Sprintf("Delete every occurrence of “%s”?", it.Title()), func() {
			m.change(func(c *calendar.Calendar) { c.Delete(it) })
		})
	}
}

// movable reports whether a Row can be Moved, and says why in the status
// line if it can't.
func (m *Model) movable(it calendar.Row) bool {
	err := it.Movable()
	if err == nil {
		return true
	}
	m.status = sentence(err)
	if errors.Is(err, calendar.ErrRecurringMove) {
		m.status += "; edit its weekdays with e"
	}
	return false
}

// move puts a Row on day and says so, or says why it can't be Moved.
func (m *Model) move(it calendar.Row, day calendar.Date) {
	if !m.movable(it) {
		return
	}
	if m.change(func(c *calendar.Calendar) { c.Move(it, day) }) {
		m.status = fmt.Sprintf("Moved “%s” to %s", it.Title(), day.Format("Mon 2 Jan"))
	}
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
		m.mode = modeNormal
		m.move(m.target, m.selected)
	case "esc", "q":
		m.mode = modeNormal
		m.selectDay(m.moveFrom)
	}
}
