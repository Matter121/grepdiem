// Tuitodo is a terminal calendar with Events and Todos.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Matter121/tuitodo/calendar"
	"github.com/Matter121/tuitodo/store"
	"github.com/Matter121/tuitodo/ui"
)

// version is set at release time with -ldflags "-X main.version=...".
var version = ""

const usage = `Usage:
  tuitodo            open the calendar
  tuitodo today      print today's events and todos
  tuitodo --version  print the version

Data is kept in $TUITODO_DATA, or ~/.local/share/tuitodo/data.json.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "tuitodo:", err)
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
		fmt.Println("tuitodo", buildVersion())
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
	st := store.New(path)
	cal, err := st.Load()
	if err != nil {
		return err
	}
	if command == "today" {
		printToday(cal, calendar.DateOf(time.Now()))
		return nil
	}
	_, err = tea.NewProgram(ui.New(cal, st, time.Now), tea.WithAltScreen()).Run()
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

func printToday(cal *calendar.Calendar, today calendar.Date) {
	day := cal.Day(today, today)
	_, week := today.ISOWeek()
	fmt.Printf("%s · week %d\n", today.Format("Mon 2 Jan 2006"), week)
	if len(day.Events) == 0 && len(day.Todos) == 0 {
		fmt.Println("\nNothing today.")
		return
	}
	if len(day.Events) > 0 {
		fmt.Println("\nEvents")
		for _, e := range day.Events {
			when := "all day"
			if e.Start != "" {
				when = calendar.TimeRange(e.Start, e.End)
			}
			fmt.Printf("  %-11s  %s\n", when, e.Title)
		}
	}
	if len(day.Todos) > 0 {
		fmt.Println("\nTodos")
		for _, t := range day.Todos {
			switch {
			case !t.Done.IsZero():
				fmt.Printf("  [x] %s\n", t.Title)
			case t.Overdue:
				fmt.Printf("  [ ] %s (from %s)\n", t.Title, t.Date.Format("Mon 2 Jan"))
			default:
				fmt.Printf("  [ ] %s\n", t.Title)
			}
		}
	}
}
