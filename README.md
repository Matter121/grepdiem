# grepdiem

A small terminal calendar with events and todos.

- Month grid with a panel for the selected day
- Weekly recurring events in one line: `Study tue thu 13-15 until w50`
- Todos per day; unfinished ones roll over to today as overdue
- One JSON file, no database, no config

## Install

```sh
go install github.com/Matter121/grepdiem/cmd/grepdiem@latest
```

Or build from a checkout:

```sh
go build -ldflags="-s -w" -trimpath -o grepdiem ./cmd/grepdiem
```

## Use

```sh
grepdiem            # open the calendar
grepdiem today      # print today's events and todos
grepdiem --version
```

Press `?` inside the app for all keys. The essentials:

| Key | Action |
|---|---|
| arrows / `hjkl` | move between days |
| `a` / `o` | add an event / a todo on the selected day |
| `enter` | open the selected day |
| `space` `e` `d` `m` `c` | tick, edit, delete, move, recolour |
| `g`, `t` | go to a date, week, month (`dec`) or year (`2028`), today |
| shift+arrows / `HJKL` | change month (left / right) or year (up / down) |
| `u` | undo the last change |
| `v` | switch between the compact view and the full view, where every day lists its events and todos |

### Adding events

Type a title, then any of these at the end of the line:

| Line | Meaning |
|---|---|
| `Dentist 10:30` | one event at 10:30 |
| `Lecture 9-11` | with an end time |
| `Study tue thu until w50` | every Tuesday and Thursday through ISO week 50 |
| `Work mon-sat 9-17` | every day from Monday to Saturday |
| `Gym mon until 13 dec` | also `until week 50`, `until 2026-12-13`, `until 261213`, `until 13/12`, `until 13.12.2026`, `until dec 13`, `until dec` (the first of December), or leave the end out |
| `Gym mon u 13 dec` | `u` is short for `until` |
| `"Apollo 13" fri 20` | quotes keep the title exactly as typed |

A preview under the line shows how it was understood before you press Enter.

## Data

Everything is stored in `~/.local/share/grepdiem/data.json` (or
`$XDG_DATA_HOME/grepdiem/data.json`). Set `GREPDIEM_DATA` to use another file.

## Development

```sh
go test ./...
```

`GLOSSARY.md` defines the project's terms; `docs/adr/` records design decisions.

## License

MIT
