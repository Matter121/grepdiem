<img width="439" height="127" alt="image" src="https://github.com/user-attachments/assets/0a1aa77b-e4b9-40bc-aa3d-43a7690f97b5" />




# grepdiem

A small terminal calendar with events and tasks.

- Month grid with a panel for the selected day
- Weekly recurring events in one line: `Study tue thu 13-15 until w50`
- Tasks per day; unfinished ones roll over to today as overdue
- Subtasks in the same line: `Pack s passport, charger, socks`
- One JSON file, no database, no config


## Install

On Linux and macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/Matter121/grepdiem/main/install.sh | sh
```

This puts the latest release in `~/.local/bin`; set `GREPDIEM_BIN` to install
somewhere else. Run it again to update. On Windows, download the zip from the
[releases page](https://github.com/Matter121/grepdiem/releases/latest) and put
`grepdiem.exe` somewhere on your `PATH`.

With Go installed:

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
grepdiem today      # print today's events and tasks
grepdiem --version
```

Press `?` inside the app for all keys. The essentials:

| Key | Action |
|---|---|
| arrows / `hjkl` | move between days |
| `a` / `o` | add an event / a task on the selected day |
| `enter` | open the selected day |
| `space` `e` `d` `m` `c` | tick, edit, delete, move, recolour |
| `>` | move a task to the next day |
| `g`, `t` | go to a date, week, month (`dec`) or year (`2028`), today |
| shift+arrows / `HJKL` | change month (left / right) or year (up / down) |
| `u` | undo the last change |
| `v` | switch between the compact view and the full view, where every day lists its events and tasks |

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

### Adding tasks

Type a title. To give the task subtasks, add `subtasks` (or `s`) and list them
with commas between:

| Line | Meaning |
|---|---|
| `Buy milk` | one task |
| `Pack s passport, charger, socks` | a task with three subtasks |
| `Pack subtasks passport, charger` | `s` is short for `subtasks` |
| `"Season s finale" s snacks, invite Sam` | quotes keep the title exactly as typed |

Each subtask is ticked on its own and the task shows how many are done (`1/3`).
Ticking the task itself is what makes it done. Subtasks have no date: they move
and go overdue with their task.

`e` on a task reopens the whole line, so that is where subtasks are added,
removed or reordered; a subtask whose text is unchanged keeps its tick. `e` on
a subtask edits just that one. The full view shows a task with its count only.

## Data

Everything is stored in `~/.local/share/grepdiem/data.json` (or
`$XDG_DATA_HOME/grepdiem/data.json`). Set `GREPDIEM_DATA` to use another file.

## Development

```sh
go test ./...
```

`GLOSSARY.md` defines the project's terms; `docs/adr/` records design decisions.


## Notes.
This is a fully vibe-coded app with Claude. Made this project so I would learn to develop with AI and also how I want my calendar to work.
The skills I used in Claude with this project are from Matt Pocock: [Github]([https://github.com/mattpocock/skills)

As noted before this only saves the events and tasks to a JSON file. I may later add syncing with external calendar apps.


## License

MIT
