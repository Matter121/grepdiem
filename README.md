# tuitodo

A small terminal calendar with events and todos.

- Month grid with a panel for the selected day
- Weekly recurring events in one line: `Study tue thu 13-15 until w50`
- Todos per day; unfinished ones roll over to today as overdue
- One JSON file, no database, no config

## Install

```sh
go install github.com/Matter121/TUICalendar/cmd/tuitodo@latest
```

Or build from a checkout:

```sh
go build -ldflags="-s -w" -trimpath -o tuitodo ./cmd/tuitodo
```

## Use

```sh
tuitodo            # open the calendar
tuitodo today      # print today's events and todos
tuitodo --version
```

Press `?` inside the app for all keys. The essentials:

| Key | Action |
|---|---|
| arrows / `hjkl` | move between days |
| `a` / `o` | add an event / a todo on the selected day |
| `enter` | open the selected day |
| `space` `e` `d` `m` `c` | tick, edit, delete, move, recolour |
| `g`, `t`, `[` `]` | go to a date or week, today, change month |
| `u` | undo the last change |

### Adding events

Type a title, then any of these at the end of the line:

| Line | Meaning |
|---|---|
| `Dentist 10:30` | one event at 10:30 |
| `Lecture 9-11` | with an end time |
| `Study tue thu until w50` | every Tuesday and Thursday through ISO week 50 |
| `Gym mon until 13 dec` | also `until week 50`, `until 2026-12-13`, or leave the end out |
| `"Apollo 13" fri 20` | quotes keep the title exactly as typed |

A preview under the line shows how it was understood before you press Enter.

## Data

Everything is stored in `~/.local/share/tuitodo/data.json` (or
`$XDG_DATA_HOME/tuitodo/data.json`). Set `TUITODO_DATA` to use another file.

## Development

```sh
go test ./...
```

`GLOSSARY.md` defines the project's terms; `docs/adr/` records design decisions.

## License

MIT
