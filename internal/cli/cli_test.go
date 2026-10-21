package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

func TestReportErrorShowsTheContext(t *testing.T) {
	var buf bytes.Buffer
	reportError(&buf, &parser.Error{
		Kind:   parser.KindSyntax,
		Msg:    "expected an integer, got \"=\"",
		Line:   12,
		Col:    31,
		Path:   "data/functions/demo/f.mcfunction",
		Source: "scoreboard players set @a obj = 5",
		Usage:  "/scoreboard players set <player> <objective> <score>",
	}, options{})

	out := buf.String()
	for _, want := range []string{
		"error[Syntax]: data/functions/demo/f.mcfunction:12:31",
		"  scoreboard players set @a obj = 5",
		"  " + strings.Repeat(" ", 30) + "^",
		"  expected an integer, got \"=\"",
		"  usage: /scoreboard players set <player> <objective> <score>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestReportErrorPlainIsOneLine(t *testing.T) {
	var buf bytes.Buffer
	reportError(&buf, &parser.Error{
		Kind:   parser.KindSyntax,
		Msg:    "expected an integer",
		Line:   12,
		Col:    31,
		Path:   "data/functions/demo/f.mcfunction",
		Source: "scoreboard players set @a obj = 5",
		Usage:  "/scoreboard players set <player> <objective> <score>",
	}, options{plainErrors: true})

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Errorf("plain output should be one line, got:\n%s", out)
	}
	if !strings.Contains(out, "error[Syntax]: 12:31: expected an integer") {
		t.Errorf("unexpected plain output:\n%s", out)
	}
	if strings.Contains(out, "usage:") || strings.Contains(out, "^") {
		t.Errorf("plain output should not carry the extra context:\n%s", out)
	}
}

func TestReportWarningUsesTheSameLayout(t *testing.T) {
	var buf bytes.Buffer
	reportWarning(&buf, &exec.FunctionLoadError{
		ID:     "demo:f",
		Path:   "data/functions/demo/f.mcfunction",
		Line:   3,
		Msg:    "Unknown or invalid command 'notacommand'",
		Source: "notacommand foo",
		Hint:   "did you mean 'scoreboard'?",
	}, options{})

	out := buf.String()
	for _, want := range []string{
		"warning[Unknown]: data/functions/demo/f.mcfunction:3:1",
		"  notacommand foo",
		"  hint: did you mean 'scoreboard'?",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

// TestReportErrorJSONIsOneObject covers the machine readable form.
//
// The point of it is that a script does not have to parse the position out
// of a line of text, so every problem is one object with the same fields.
func TestReportErrorJSONIsOneObject(t *testing.T) {
	var buf bytes.Buffer
	reportError(&buf, &parser.Error{
		Kind:   parser.KindSyntax,
		Msg:    "expected an integer, got \"=\"",
		Line:   12,
		Col:    31,
		Path:   "data/functions/demo/f.mcfunction",
		Source: "scoreboard players set @a obj = 5",
		Usage:  "/scoreboard players set <player> <objective> <score>",
	}, options{json: true})

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Errorf("the JSON output should be one line, got:\n%s", out)
	}
	var e jsonError
	if err := json.Unmarshal([]byte(out), &e); err != nil {
		t.Fatalf("the output is not JSON: %v\n%s", err, out)
	}
	if e.Level != "error" || e.Kind != "Syntax" {
		t.Errorf("level/kind = %q/%q, want error/Syntax", e.Level, e.Kind)
	}
	if e.Path != "data/functions/demo/f.mcfunction" || e.Line != 12 || e.Col != 31 {
		t.Errorf("the position is wrong: %+v", e)
	}
	if e.Msg != "expected an integer, got \"=\"" {
		t.Errorf("msg = %q", e.Msg)
	}
	if e.Usage != "/scoreboard players set <player> <objective> <score>" {
		t.Errorf("usage = %q", e.Usage)
	}
}

// TestRunAsPlayer covers the "--as" switch.
//
// The player is the listener of the commands, so the output of a `say`
// carries the player's name instead of "Server" — the same as on a real
// server, where a player runs the function.
func TestRunAsPlayer(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "functions", "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hello.mcfunction"), []byte("say hi\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"run", "--as", "Steve", root, "demo:hello"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run --as: exit %d\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[Steve] hi") {
		t.Errorf("the output does not name the player:\n%s", stdout.String())
	}

	// Without the switch the console is the source.
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"run", root, "demo:hello"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run: exit %d\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[Server] hi") {
		t.Errorf("the output does not name the console:\n%s", stdout.String())
	}
}

// TestEverySubcommandHasHelp covers that "-h" explains a subcommand
// instead of being taken for an argument.
func TestEverySubcommandHasHelp(t *testing.T) {
	for _, args := range [][]string{
		{"check", "-h"},
		{"run", "-h"},
		{"tick", "-h"},
		{"commands", "-h"},
		{"difftest", "-h"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 0 {
			t.Errorf("%v: exit %d\n%s", args, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "Usage:") {
			t.Errorf("%v: the help does not show a usage line:\n%s", args, stdout.String())
		}
	}
}

// TestDifftestList covers the listing of the cases.
func TestDifftestList(t *testing.T) {
	dir := t.TempDir()
	body := "name: a case\nrun:\n  say hi\n"
	if err := os.WriteFile(filepath.Join(dir, "a.difftest"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"difftest", "list", dir}, &stdout, &stderr); code != 0 {
		t.Fatalf("difftest list: exit %d\n%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "a case") {
		t.Errorf("the listing does not name the case:\n%s", out)
	}
	if !strings.Contains(out, "no golden") {
		t.Errorf("the listing does not say that the golden is missing:\n%s", out)
	}
}
