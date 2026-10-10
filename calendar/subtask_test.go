package calendar

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestParseTaskSplitsOffSubtasks(t *testing.T) {
	cases := []struct {
		line string
		want TaskSpec
	}{
		{"Buy milk", TaskSpec{Title: "Buy milk"}},
		{"Pack s passport, charger, socks", TaskSpec{"Pack", []string{"passport", "charger", "socks"}}},
		{"Pack for trip subtasks passport,phone charger", TaskSpec{"Pack for trip", []string{"passport", "phone charger"}}},
		{"Pack S passport", TaskSpec{"Pack", []string{"passport"}}},
		// The first keyword splits the line; a later one belongs to a Subtask.
		{"Pack s size s shirts, socks", TaskSpec{"Pack", []string{"size s shirts", "socks"}}},
		// A keyword needs a word before it and text after it.
		{"Grade s", TaskSpec{Title: "Grade s"}},
		{"Grade s ,", TaskSpec{Title: "Grade s ,"}},
		{"s passport, charger", TaskSpec{Title: "s passport, charger"}},
		{"Pack s a,, b ,", TaskSpec{"Pack", []string{"a", "b"}}},
		{`"Season s finale" s buy snacks, invite Sam`, TaskSpec{"Season s finale", []string{"buy snacks", "invite Sam"}}},
		{`"Season s finale"`, TaskSpec{Title: "Season s finale"}},
	}
	for _, c := range cases {
		if got := ParseTask(c.line); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: got %+v, want %+v", c.line, got, c.want)
		}
	}
}

func TestTaskLineRecreatesTheTask(t *testing.T) {
	cases := []struct {
		spec TaskSpec
		want string
	}{
		{TaskSpec{Title: "Buy milk"}, "Buy milk"},
		{TaskSpec{"Pack", []string{"passport", "phone charger"}}, "Pack s passport, phone charger"},
		{TaskSpec{Title: "Grade s papers"}, `"Grade s papers"`},
		{TaskSpec{"Season s finale", []string{"buy snacks"}}, `"Season s finale" s buy snacks`},
	}
	for _, c := range cases {
		got := TaskLine(c.spec)
		if got != c.want {
			t.Errorf("%+v: got %q, want %q", c.spec, got, c.want)
		}
		if back := ParseTask(got); !reflect.DeepEqual(back, c.spec) {
			t.Errorf("%q reads back as %+v, want %+v", got, back, c.spec)
		}
	}
}

func TestDescribeTaskCountsItsSubtasks(t *testing.T) {
	cases := map[string]string{
		"Buy milk":                        "Buy milk",
		"Pack s passport":                 "Pack · 1 subtask: passport",
		"Pack s passport, charger, socks": "Pack · 3 subtasks: passport, charger, socks",
	}
	for line, want := range cases {
		if got := DescribeTask(ParseTask(line)); got != want {
			t.Errorf("%q: got %q, want %q", line, got, want)
		}
	}
}

type line struct {
	kind  RowKind
	state TaskState
	text  string
}

func lines(day Day) []line {
	var out []line
	for _, r := range day.Rows {
		text := r.Mark() + " " + r.Title()
		if count := r.Count(); count != "" {
			text += " " + count
		}
		out = append(out, line{r.Kind, r.State, text})
	}
	return out
}

func TestSubtasksAreRowsUnderTheirTask(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger", "socks")
	c.AddTask(today, "Buy milk")
	c.ToggleSubtask(pack.ID, 0)

	want := []line{
		{TaskRow, Open, "[ ] Pack 1/3"},
		{SubtaskRow, Done, "[x] passport"},
		{SubtaskRow, Open, "[ ] charger"},
		{SubtaskRow, Open, "[ ] socks"},
		{TaskRow, Open, "[ ] Buy milk"},
	}
	day := c.Day(today, today)
	if got := lines(day); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if r := day.Rows[2]; r.Task.ID != pack.ID || r.Sub != 1 {
		t.Errorf("row %+v should stand for Subtask 1 of Task %d", r, pack.ID)
	}
	if top := day.WithoutSubtasks(); len(top) != 2 || top[1].Title() != "Buy milk" {
		t.Errorf("without Subtasks: %+v", top)
	}

	c.ToggleSubtask(pack.ID, 0)
	if got := lines(c.Day(today, today))[:2]; got[0].text != "[ ] Pack 0/3" || got[1].state != Open {
		t.Errorf("after unticking: %+v", got)
	}
}

func TestTickingEverySubtaskLeavesTheTaskOpen(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger")
	c.ToggleSubtask(pack.ID, 0)
	c.ToggleSubtask(pack.ID, 1)

	if got := lines(c.Day(today, today))[0]; got != (line{TaskRow, Open, "[ ] Pack 2/2"}) {
		t.Errorf("got %+v", got)
	}
}

func TestDoneTaskHidesItsSubtasksUntilReopened(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger")
	c.ToggleSubtask(pack.ID, 1)
	c.ToggleTask(pack.ID, today)

	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, []line{{TaskRow, Done, "[x] Pack 1/2"}}) {
		t.Errorf("done: %+v", got)
	}
	c.ToggleTask(pack.ID, today)
	want := []line{{TaskRow, Open, "[ ] Pack 1/2"}, {SubtaskRow, Open, "[ ] passport"}, {SubtaskRow, Done, "[x] charger"}}
	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, want) {
		t.Errorf("reopened: got %+v, want %+v", got, want)
	}
}

func TestOverdueTaskBringsItsSubtasksToToday(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	c.AddTask(d("2026-10-07"), "Pack", "passport")
	c.AddTask(d("2026-10-06"), "Rent")

	rows := c.Day(today, today).Rows
	want := []line{{TaskRow, Overdue, "[ ] Rent"}, {TaskRow, Overdue, "[ ] Pack 0/1"}, {SubtaskRow, Open, "[ ] passport"}}
	if got := lines(Day{Rows: rows}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if rows[1].Note() != "overdue from Wed 7 Oct" || rows[2].Note() != "" {
		t.Errorf("notes %q and %q: only the Task says it is Overdue", rows[1].Note(), rows[2].Note())
	}
}

func TestUpdateTaskKeepsTheTicksOfSubtasksItStillHas(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger", "socks")
	c.ToggleSubtask(pack.ID, 0)
	c.ToggleSubtask(pack.ID, 2)

	c.UpdateTask(pack.ID, "Pack bag", "socks", "wallet", "passports")
	want := []line{
		{TaskRow, Open, "[ ] Pack bag 1/3"},
		{SubtaskRow, Done, "[x] socks"},
		{SubtaskRow, Open, "[ ] wallet"},
		{SubtaskRow, Open, "[ ] passports"},
	}
	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestUpdateTaskMatchesSubtasksWithTheSameTitleInOrder(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	task := c.AddTask(today, "Laundry", "load", "load")
	c.ToggleSubtask(task.ID, 0)

	c.UpdateTask(task.ID, "Laundry", "load", "load", "load")
	var ticks []bool
	for _, s := range c.Tasks[0].Subtasks {
		ticks = append(ticks, s.Done)
	}
	if !reflect.DeepEqual(ticks, []bool{true, false, false}) {
		t.Errorf("ticks %v, want only the first", ticks)
	}
}

func TestSubtaskIsRenamedAndDeletedOnItsOwn(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger", "socks")
	c.ToggleSubtask(pack.ID, 1)

	c.RenameSubtask(pack.ID, 1, "phone charger")
	c.DeleteSubtask(pack.ID, 0)
	c.DeleteSubtask(pack.ID, 7) // no such Subtask
	want := []line{{TaskRow, Open, "[ ] Pack 1/2"}, {SubtaskRow, Done, "[x] phone charger"}, {SubtaskRow, Open, "[ ] socks"}}
	if got := lines(c.Day(today, today)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if spec := c.Tasks[0].Spec(); !reflect.DeepEqual(spec, TaskSpec{"Pack", []string{"phone charger", "socks"}}) {
		t.Errorf("spec %+v", spec)
	}
}

func TestSubtasksSurviveJSONAndCloneSharesNone(t *testing.T) {
	c := New()
	today := d("2026-10-09")
	pack := c.AddTask(today, "Pack", "passport", "charger")
	c.AddTask(today, "Buy milk")
	c.ToggleSubtask(pack.ID, 0)

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if want := `"subtasks":[{"title":"passport","done":true},{"title":"charger"}]`; !strings.Contains(string(data), want) {
		t.Errorf("JSON lacks %s:\n%s", want, data)
	}
	if n := strings.Count(string(data), `"subtasks"`); n != 1 {
		t.Errorf("a Task without Subtasks should have no subtasks field, found %d", n)
	}
	back := New()
	if err := json.Unmarshal(data, back); err != nil {
		t.Fatal(err)
	}
	if a, b := lines(c.Day(today, today)), lines(back.Day(today, today)); !reflect.DeepEqual(a, b) {
		t.Errorf("differs after JSON: %+v vs %+v", a, b)
	}

	before := c.Clone()
	c.ToggleSubtask(pack.ID, 1)
	if before.Tasks[0].Subtasks[1].Done {
		t.Error("ticking a Subtask changed the clone")
	}
}
