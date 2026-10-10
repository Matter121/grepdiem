// Grepdiem is a terminal calendar with Events and Tasks.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Matter121/grepdiem/calendar"
	"github.com/Matter121/grepdiem/store"
	"github.com/Matter121/grepdiem/ui"
)

// version is set at release time with -ldflags "-X main.version=...".
var version = ""

const usage = `Usage:
  grepdiem            open the calendar
  grepdiem today      print today's events and tasks
  grepdiem --version  print the version

Data is kept in $GREPDIEM_DATA, or ~/.local/share/grepdiem/data.json.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "grepdiem:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "--version", "-v", "version":
		fmt.Println("grepdiem", buildVersion())
		return nil
	case "--help", "-h", "help":
		fmt.Print(usage)
		return nil
	case "", "today":
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", command)
	}

	path, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.Open(path)
	if err != nil {
		return err
	}
	if command == "today" {
		printToday(os.Stdout, st.Calendar(), calendar.DateOf(time.Now()))
		return nil
	}
	_, err = tea.NewProgram(ui.New(st, time.Now), tea.WithAltScreen()).Run()
	return err
}

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func printToday(w io.Writer, cal *calendar.Calendar, today calendar.Date) {
	rows := cal.Day(today, today).Rows
	_, week := today.ISOWeek()
	fmt.Fprintf(w, "%s · week %d\n", today.Format("Mon 2 Jan 2006"), week)
	if len(rows) == 0 {
		fmt.Fprintln(w, "\nNothing today.")
		return
	}
	for i, r := range rows {
		switch {
		case r.Kind == calendar.EventRow && i == 0:
			fmt.Fprintln(w, "\nEvents")
		case r.Kind != calendar.EventRow && (i == 0 || rows[i-1].Kind == calendar.EventRow):
			fmt.Fprintln(w, "\nTasks")
		}
		if r.Kind == calendar.EventRow {
			when := r.Time()
			if when == "" {
				when = "all day"
			}
			fmt.Fprintf(w, "  %-11s  %s\n", when, r.Title())
			continue
		}
		line := "  " + r.Mark() + " " + r.Title()
		if r.Kind == calendar.SubtaskRow {
			line = "    " + line
		}
		if count := r.Count(); count != "" {
			line += "  " + count
		}
		if note := r.Note(); note != "" {
			line += " (" + note + ")"
		}
		fmt.Fprintln(w, line)
	}
}
