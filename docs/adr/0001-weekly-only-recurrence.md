# Weekly-only recurrence in our own format, not iCalendar RRULE

A Recurring Event is stored as a set of weekdays, a start date, an optional end date and a list of Skips, and that is the only repetition the app supports. We chose this over the iCalendar RRULE standard because the one case we care about (a weekly timetable until a given week) stays trivial to enter, store and test, and it avoids pulling in an RRULE library that would grow the binary.

## Consequences

Importing from or exporting to iCal/Google Calendar is not a straight mapping and would need a conversion layer. Monthly, yearly and every-N-weeks repetition cannot be expressed without changing the data format.
