package ui

import (
	"slices"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lineInput struct {
	runes []rune
	pos   int
}

func (l *lineInput) set(s string) {
	l.runes = []rune(s)
	l.pos = len(l.runes)
}

func (l *lineInput) String() string { return string(l.runes) }

func (l *lineInput) handle(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace:
		if msg.Alt {
			return
		}
		l.runes = slices.Insert(l.runes, l.pos, msg.Runes...)
		l.pos += len(msg.Runes)
	case tea.KeyBackspace:
		if l.pos > 0 {
			l.runes = slices.Delete(l.runes, l.pos-1, l.pos)
			l.pos--
		}
	case tea.KeyDelete:
		if l.pos < len(l.runes) {
			l.runes = slices.Delete(l.runes, l.pos, l.pos+1)
		}
	case tea.KeyLeft:
		l.pos = max(l.pos-1, 0)
	case tea.KeyRight:
		l.pos = min(l.pos+1, len(l.runes))
	case tea.KeyHome, tea.KeyCtrlA:
		l.pos = 0
	case tea.KeyEnd, tea.KeyCtrlE:
		l.pos = len(l.runes)
	case tea.KeyCtrlU:
		l.runes = l.runes[l.pos:]
		l.pos = 0
	case tea.KeyCtrlW:
		start := l.pos
		for start > 0 && unicode.IsSpace(l.runes[start-1]) {
			start--
		}
		for start > 0 && !unicode.IsSpace(l.runes[start-1]) {
			start--
		}
		l.runes = slices.Delete(l.runes, start, l.pos)
		l.pos = start
	}
}

// view draws the text with a block cursor, scrolled so the cursor stays in width.
func (l *lineInput) view(width int) string {
	width = max(width, 4)
	start := max(0, l.pos-width+1)
	end := min(len(l.runes), start+width-1)
	var b strings.Builder
	b.WriteString(string(l.runes[start:l.pos]))
	cursor := " "
	if l.pos < len(l.runes) {
		cursor = string(l.runes[l.pos])
	}
	b.WriteString(lipgloss.NewStyle().Reverse(true).Render(cursor))
	if l.pos+1 <= end {
		b.WriteString(string(l.runes[l.pos+1 : end]))
	}
	return b.String()
}
