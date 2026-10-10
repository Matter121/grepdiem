# A Task's tick is independent of its Subtasks

Ticking Subtasks only changes the count shown beside the Task; the Task is Done only when it is ticked itself, and it can be ticked while Subtasks are still open. We chose this over linking the two (the last Subtask ticks the Task, unticking one reopens it) and over deriving the Task's state from its Subtasks, because Done is what moves a Task to the day it was ticked and stops it being Overdue, and that should not happen, or be reversed, as a side effect of ticking something else.

## Consequences

A Task with every Subtask ticked stays open, and Overdue, until it is ticked. A Done Task can hold unticked Subtasks; they are hidden while it is Done and come back as they were if it is reopened. Subtasks are matched by title when a Task's line is edited, so they carry no IDs in the data file.
