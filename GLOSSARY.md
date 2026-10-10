# grepdiem

A personal terminal calendar that shows what happens on each day and what needs doing.

## Language

**Event**:
Something that happens on a date, optionally at a time of day. It cannot be completed.
_Avoid_: Marking, appointment, entry

**Recurring Event**:
An Event that repeats weekly on one or more weekdays, from a start date until an optional end.
_Avoid_: Series, repeat, schedule

**Occurrence**:
One dated instance of a Recurring Event.
_Avoid_: Instance

**Skip**:
The removal of a single Occurrence from a Recurring Event while the rest remain.
_Avoid_: Exception, cancellation

**All-day Event**:
An Event with no time of day.

**Task**:
Something to be done and ticked off, attached to a date.
_Avoid_: Todo, item

**Subtask**:
One step of a Task, with its own tick and no date of its own; it goes wherever its Task goes. Ticking Subtasks never makes the Task Done.
_Avoid_: Step, checklist item, child

**Overdue**:
The state of an unfinished Task whose date has passed; it is shown under today until ticked or deleted.
_Avoid_: Late, rolled over

**Done**:
The state of a Task that has been ticked; it is shown on the day it was ticked, not the day it was attached to, as one Row without its Subtasks. A ticked Subtask is also Done, which changes nothing about its Task.
_Avoid_: Completed, finished, checked

**Row**:
One line in a day's listing, standing for one Event, one Task or one Subtask.
_Avoid_: Item, entry, line

**Quick-add**:
The single line of text from which an Event, or a Task with its Subtasks, is created.
_Avoid_: Command, prompt, form

**Marker**:
The coloured dot on a day in the month grid that stands for one Event on that day.
_Avoid_: Marking, badge, indicator

**One-off Event**:
An Event that happens on a single date and does not repeat.
_Avoid_: Single event, normal event

**Move**:
Changing the date a Task or One-off Event is attached to; a moved Task is no longer Overdue and takes its Subtasks with it.
_Avoid_: Reschedule, postpone

**Compact view**:
The small month grid with a Marker for each Event, beside a panel that lists the selected day's Events, Tasks and Subtasks.
_Avoid_: Normal view, small view, default view

**Full view**:
The month grid filling the whole terminal, with every day listing its own Events and Tasks; a Task's Subtasks are shown only as a count.
_Avoid_: Big view, month view, expanded view
