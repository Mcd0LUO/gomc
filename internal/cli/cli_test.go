package cli

import (
	"bytes"
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
