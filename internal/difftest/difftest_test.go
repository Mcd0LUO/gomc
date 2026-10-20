package difftest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCase(t *testing.T) {
	text := `# a differential case
name: scoreboard add

setup:
  scoreboard objectives add counter dummy
  scoreboard players set Steve counter 1

run:
  scoreboard players add Steve counter 2
  say done
`
	c, err := ParseCase(text)
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	if c.Name != "scoreboard add" {
		t.Errorf("name = %q, want %q", c.Name, "scoreboard add")
	}
	if len(c.Setup) != 2 {
		t.Errorf("setup = %v, want two commands", c.Setup)
	}
	if len(c.Run) != 2 {
		t.Errorf("run = %v, want two commands", c.Run)
	}
	if c.Run[1] != "say done" {
		t.Errorf("run[1] = %q, want %q", c.Run[1], "say done")
	}
}

func TestParseCaseRejectsCommandOutsideSection(t *testing.T) {
	if _, err := ParseCase("scoreboard objectives add a dummy\n"); err == nil {
		t.Fatal("expected an error for a command outside a section")
	}
}

func TestParseCaseRequiresRunSection(t *testing.T) {
	if _, err := ParseCase("name: empty\nsetup:\n  say hi\n"); err == nil {
		t.Fatal("expected an error for a case without a run section")
	}
}

// newCaseWorld writes a world with a function and returns its root.
func newCaseWorld(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "data", "functions", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "scoreboard objectives add ticks dummy\nscoreboard players add s ticks 1\n"
	if err := os.WriteFile(filepath.Join(dir, "loop.mcfunction"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return root
}

func TestRunRecordsFeedback(t *testing.T) {
	root := newCaseWorld(t)
	c, err := ParseCase("name: feedback\nrun:\n  say hello\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	obs, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(obs.Steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(obs.Steps))
	}
	if obs.Steps[0].Err != "" {
		t.Errorf("error = %q, want none", obs.Steps[0].Err)
	}
	if len(obs.Steps[0].Feedback) == 0 {
		t.Error("the say command produced no feedback")
	}
}

func TestRunRecordsErrors(t *testing.T) {
	root := newCaseWorld(t)
	c, err := ParseCase("name: errors\nrun:\n  nosuchcommand foo\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	obs, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if obs.Steps[0].Err == "" {
		t.Fatal("the unknown command produced no error")
	}
	if !strings.Contains(obs.Steps[0].Err, "nosuchcommand") {
		t.Errorf("error = %q, want it to name the command", obs.Steps[0].Err)
	}
}

// TestRunIsDeterministic guards the property the whole layer relies on.
//
// A golden file is only useful when two runs of the same case produce the
// same text.
func TestRunIsDeterministic(t *testing.T) {
	root := newCaseWorld(t)
	c, err := ParseCase("name: deterministic\nsetup:\n  scoreboard objectives add counter dummy\nrun:\n  scoreboard players set Steve counter 1\n  scoreboard players add Steve counter 2\n  summon armor_stand ~ ~ ~\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	first, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	second, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if first.Render(ScopeServer) != second.Render(ScopeServer) {
		t.Errorf("the two runs differ:\n%s\n%s", first.Render(ScopeServer), second.Render(ScopeServer))
	}
}

// TestRunIsDeterministicWithRandom guards the same property for the parts
// that are randomised.
//
// The `@r` selector and the default duration of `/weather` take from the
// random source of the world, so they depend on the order of the commands.
// What must not depend on it is the run before: a case that runs after
// another case has to see the same random source.
//
// (The random source is reset by `exec.NewEngine`, so every case starts
// over.)
func TestRunIsDeterministicWithRandom(t *testing.T) {
	root := newCaseWorld(t)
	caseText := "name: random\nsetup:\n  summon zombie 0 4 0 {CustomName:\"A\"}\n  summon zombie 1 4 0 {CustomName:\"B\"}\n  summon zombie 2 4 0 {CustomName:\"C\"}\nrun:\n  execute @r[type=zombie] ~ ~ ~ say picked\n"
	c, err := ParseCase(caseText)
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	first, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// A run in between takes from the same random source.
	if _, err := Run(root, c); err != nil {
		t.Fatalf("Run: %v", err)
	}
	second, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if first.Render(ScopeServer) != second.Render(ScopeServer) {
		t.Errorf("the two runs differ:\n%s\n%s", first.Render(ScopeServer), second.Render(ScopeServer))
	}
}

func TestRenderIsStable(t *testing.T) {
	root := newCaseWorld(t)
	c, err := ParseCase("name: render\nrun:\n  say hello\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	obs, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := obs.Render(ScopeServer)
	if !strings.Contains(got, "step 1: say hello") {
		t.Errorf("render = %q, want it to name the step", got)
	}
	if !strings.Contains(got, "feedback:") {
		t.Errorf("render = %q, want it to include the feedback", got)
	}
}

func TestDiffReportsDifferences(t *testing.T) {
	diff := Diff("a\nb\nc\n", "a\nx\nc\n")
	if len(diff) != 2 {
		t.Fatalf("diff = %v, want two lines", diff)
	}
	if diff[0] != "- b" || diff[1] != "+ x" {
		t.Errorf("diff = %v, want [- b, + x]", diff)
	}
	if len(Diff("a\n", "a\n")) != 0 {
		t.Error("identical texts produced a diff")
	}
}

func TestProbeIncludesEveryCommand(t *testing.T) {
	c, err := ParseCase("name: probe\nsetup:\n  say setup\nrun:\n  say one\n  say two\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	probe := Probe(c)
	for _, want := range []string{"say setup", "say one", "say two", "## step 1", "## step 2"} {
		if !strings.Contains(probe, want) {
			t.Errorf("the probe is missing %q:\n%s", want, probe)
		}
	}
}

// TestRenderStateShowsTheWorld covers the state rendering.
func TestRenderStateShowsTheWorld(t *testing.T) {
	root := newCaseWorld(t)
	c, err := ParseCase("name: state\nrun:\n  scoreboard objectives add counter dummy\n  scoreboard players set Steve counter 3\n  setblock 0 0 0 stone\n")
	if err != nil {
		t.Fatalf("ParseCase: %v", err)
	}
	obs, err := Run(root, c)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	full := obs.Render(ScopeFull)
	for _, want := range []string{"counter/Steve = 3", "0,0,0 = stone 0"} {
		if !strings.Contains(full, want) {
			t.Errorf("the full render is missing %q:\n%s", want, full)
		}
	}
	// The server scope leaves the state out, because a real server cannot
	// report it.
	server := obs.Render(ScopeServer)
	if strings.Contains(server, "counter/Steve") {
		t.Errorf("the server render should not contain the state:\n%s", server)
	}
}
