// Package ui is the terminal interface.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Matter121/TUICalendar/calendar"
	"github.com/Matter121/TUICalendar/store"
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
	inputAddTodo
	inputEditEvent
	inputEditTodo
	inputGoto
)

// item is one row of the day panel: an Event or a Todo.
type item struct {
	isTodo bool
	event  calendar.Event
	todo   calendar.DayTodo
}

func (it item) title() string {
	if it.isTodo {
		return it.todo.Title
	}
	return it.event.Title
}

type Model struct {
	cal   *calendar.Calendar
	store *store.Store
	now   func() time.Time

	selected calendar.Date
	focus    focus
	cursor   int // index into the day panel's items

	mode      mode
	input     lineInput
	inputKind inputKind
	target    item          // what is being edited, deleted or moved
	moveFrom  calendar.Date // where to return if a move is cancelled
	question  string        // modeConfirm
	onYes     func()

	helpScroll int

	undo   *calendar.Calendar
	status string
	width  int
	height int
}

// New builds the UI on a loaded Calendar. now is the clock "today" comes from.
func New(cal *calendar.Calendar, st *store.Store, now func() time.Time) *Model {
	return &Model{cal: cal, store: st, now: now, selected: calendar.DateOf(now())}
}

type tickMsg struct{}

// tick redraws once a minute so "today" moves on at midnight.
func tick() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) Init() tea.Cmd { return tick() }

func (m *Model) today() calendar.Date { return calendar.DateOf(m.now()) }

func (m *Model) items() []item {
	day := m.cal.Day(m.selected, m.today())
	var items []item
	for _, e := range day.Events {
		items = append(items, item{event: e})
	}
	for _, t := range day.Todos {
		items = append(items, item{isTodo: true, todo: t})
	}
	return items
}

// change applies an edit and saves it. If the file changed on disk since it
// was read, the edit is applied on top of what is there now.
func (m *Model) change(edit func(c *calendar.Calendar)) {
	if m.store.Changed() {
		fresh, err := m.store.Load()
		if err != nil {
			m.status = "Could not reload: " + err.Error()
			return
		}
		m.cal = fresh
	}
	m.undo = m.cal.Clone()
	edit(m.cal)
	m.save()
}

func (m *Model) save() {
	if err := m.store.Save(m.cal); err != nil {
		m.status = "Could not save: " + err.Error()
	}
}

func (m *Model) selectDay(d calendar.Date) {
	m.selected = d
	m.cursor = 0
}

// settle keeps the cursor and focus valid after the day's items change.
func (m *Model) settle() {
	n := len(m.items())
	m.cursor = max(0, min(m.cursor, n-1))
	if n == 0 && m.mode != modeMove {
		m.focus = focusGrid
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
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
	case "[":
		m.selectDay(m.selected.AddMonths(-1))
	case "]":
		m.selectDay(m.selected.AddMonths(1))
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
		m.openInput(inputAddTodo, "")
	case "g":
		m.openInput(inputGoto, "")
	case "u":
		m.undoLast()
	case "tab":
		if m.focus == focusGrid && len(m.items()) > 0 {
			m.focus = focusPanel
		} else {
			m.focus = focusGrid
		}
	case "enter":
		if len(m.items()) > 0 {
			m.focus = focusPanel
		}
	case "esc":
		m.focus = focusGrid
	}

	if m.focus != focusPanel {
		return nil
	}
	items := m.items()
	if m.cursor >= len(items) {
		return nil
	}
	it := items[m.cursor]
	switch key {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, len(items)-1)
	case " ":
		if it.isTodo {
			today := m.today()
			m.change(func(c *calendar.Calendar) { c.ToggleTodo(it.todo.ID, today) })
		}
	case "e":
		m.target = it
		if it.isTodo {
			m.openInput(inputEditTodo, it.todo.Title)
		} else {
			m.openInput(inputEditEvent, calendar.Line(it.event.Spec()))
		}
	case "d":
		m.delete(it)
	case "c":
		if !it.isTodo {
			m.change(func(c *calendar.Calendar) { c.CycleColor(it.event.ID) })
		}
	case "m":
		m.startMove(it)
	case ">":
		if m.movable(it) && it.isTodo {
			next := m.selected.AddDays(1)
			m.change(func(c *calendar.Calendar) { c.MoveTodo(it.todo.ID, next) })
			m.status = fmt.Sprintf("Moved “%s” to %s", it.title(), next.Format("Mon 2 Jan"))
		}
	}
	return nil
}

func (m *Model) undoLast() {
	if m.undo == nil {
		m.status = "Nothing to undo"
		return
	}
	m.cal, m.undo = m.undo, m.cal
	m.save()
	m.status = "Undone (u again to redo)"
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
	case inputAddTodo:
		day := m.selected
		m.change(func(c *calendar.Calendar) { c.AddTodo(day, text) })
	case inputEditTodo:
		id := m.target.todo.ID
		m.change(func(c *calendar.Calendar) { c.RenameTodo(id, text) })
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
		m.change(func(c *calendar.Calendar) { added = c.AddEvent(day, spec) })
		m.status = addedStatus(added, day)
	case inputEditEvent:
		spec := calendar.ParseEvent(text, m.selected)
		if spec.Title == "" {
			m.status = "An event needs a title"
			return
		}
		old := m.target.event
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

func (m *Model) delete(it item) {
	switch {
	case it.isTodo:
		m.change(func(c *calendar.Calendar) { c.DeleteTodo(it.todo.ID) })
	case it.event.Recurring():
		m.target = it
		m.mode = modeDeleteChoice
	default:
		m.change(func(c *calendar.Calendar) { c.DeleteEvent(it.event.ID) })
	}
}

func (m *Model) updateDeleteChoice(msg tea.KeyMsg) {
	m.mode = modeNormal
	e, day := m.target.event, m.selected
	switch msg.String() {
	case "t":
		m.change(func(c *calendar.Calendar) { c.Skip(e.ID, day) })
		m.status = fmt.Sprintf("Skipped “%s” on %s", e.Title, day.Format("Mon 2 Jan"))
	case "a":
		m.ask(fmt.Sprintf("Delete every occurrence of “%s”?", e.Title), func() {
			m.change(func(c *calendar.Calendar) { c.DeleteEvent(e.ID) })
		})
	}
}

func (m *Model) movable(it item) bool {
	switch {
	case !it.isTodo && it.event.Recurring():
		m.status = "A recurring event can't be moved; edit its weekdays with e"
	case it.isTodo && !it.todo.Done.IsZero():
		m.status = "A done todo can't be moved"
	default:
		return true
	}
	return false
}

func (m *Model) startMove(it item) {
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
		m.change(func(c *calendar.Calendar) {
			if it.isTodo {
				c.MoveTodo(it.todo.ID, day)
			} else {
				c.MoveEvent(it.event.ID, day)
			}
		})
		m.mode = modeNormal
		m.status = fmt.Sprintf("Moved “%s” to %s", it.title(), day.Format("Mon 2 Jan"))
	case "esc", "q":
		m.mode = modeNormal
		m.selectDay(m.moveFrom)
	}
}
