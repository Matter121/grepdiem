# TUItodo

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

**Todo**:
Something to be done and ticked off, attached to a date.
_Avoid_: Task, item

**Overdue**:
The state of an unfinished Todo whose date has passed; it is shown under today until ticked or deleted.
_Avoid_: Late, rolled over

**Done**:
The state of a Todo that has been ticked; it is shown on the day it was ticked, not the day it was attached to.
_Avoid_: Completed, finished, checked

**Row**:
One line in a day's listing, standing for one Event or one Todo.
_Avoid_: Item, entry, line

**Quick-add**:
The single line of text from which an Event or Todo is created.
_Avoid_: Command, prompt, form

**Marker**:
The coloured dot on a day in the month grid that stands for one Event on that day.
_Avoid_: Marking, badge, indicator

**One-off Event**:
An Event that happens on a single date and does not repeat.
_Avoid_: Single event, normal event

**Move**:
Changing the date a Todo or One-off Event is attached to; a moved Todo is no longer Overdue.
_Avoid_: Reschedule, postpone

**Compact view**:
The small month grid with a Marker for each Event, beside a panel that lists the selected day's Events and Todos.
_Avoid_: Normal view, small view, default view

**Full view**:
The month grid filling the whole terminal, with every day listing its own Events and Todos.
_Avoid_: Big view, month view, expanded view
