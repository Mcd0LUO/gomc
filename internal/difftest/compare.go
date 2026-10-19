package difftest

import (
	"fmt"
	"strings"
)

// Diff compares two canonical observations.
//
// The result lists the lines that differ, in the order they appear in the
// got text, so the output reads like a normal diff.
func Diff(want, got string) []string {
	wantLines := splitLines(want)
	gotLines := splitLines(got)

	var out []string
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w == g {
			continue
		}
		switch {
		case w == "":
			out = append(out, fmt.Sprintf("+ %s", g))
		case g == "":
			out = append(out, fmt.Sprintf("- %s", w))
		default:
			out = append(out, fmt.Sprintf("- %s", w))
			out = append(out, fmt.Sprintf("+ %s", g))
		}
	}
	return out
}

// splitLines splits a text into lines, dropping the trailing empty one.
func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// FormatDiff renders a diff for the terminal.
func FormatDiff(diff []string) string {
	if len(diff) == 0 {
		return ""
	}
	return strings.Join(diff, "\n")
}

// Probe renders a function that reproduces a case on a real server.
//
// A 1.12.2 server cannot dump its state, but it does print the feedback of
// every command and the error message of a failing one. The probe
// therefore runs the commands with a marker before each one, so the server
// log can be lined up with the server scope of an observation.
func Probe(c *Case) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Probe for the differential case %q.\n", c.Name)
	fmt.Fprintf(&sb, "# Run this on a 1.12.2 server and compare the log with\n")
	fmt.Fprintf(&sb, "# `gomc difftest run <world> <case>`.\n")
	for _, cmd := range c.Setup {
		fmt.Fprintf(&sb, "say ## setup: %s\n", cmd)
		fmt.Fprintf(&sb, "%s\n", cmd)
	}
	for i, cmd := range c.Run {
		fmt.Fprintf(&sb, "say ## step %d: %s\n", i+1, cmd)
		fmt.Fprintf(&sb, "%s\n", cmd)
	}
	return sb.String()
}
