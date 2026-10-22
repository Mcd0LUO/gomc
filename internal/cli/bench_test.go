package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeBenchFunction writes a function into a temporary world.
func writeBenchFunction(t *testing.T, ns, name, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "data", "functions", ns)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, name+".mcfunction")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return root
}

// TestBenchMeasuresAProject covers the report of a measurement.
func TestBenchMeasuresAProject(t *testing.T) {
	root := writeBenchFunction(t, "pack", "main",
		"scoreboard objectives add counter dummy\n"+
			"scoreboard players set Steve counter 0\n"+
			"say hello\n")

	var out bytes.Buffer
	code := runBench([]string{root, "--all"}, &out, &out, options{})
	if code != 0 {
		t.Fatalf("the measurement failed with the code %d:\n%s", code, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"mode:   every function",
		"1 function(s)",
		"3 command(s)",
		"0 failure(s)",
		"by function",
		"by command",
		"pack:main",
		"scoreboard",
		"say",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the report does not contain %q:\n%s", want, text)
		}
	}
}

// TestBenchJSON covers the machine readable report.
func TestBenchJSON(t *testing.T) {
	root := writeBenchFunction(t, "pack", "main", "say hello\n")

	var out bytes.Buffer
	code := runBench([]string{root, "--all", "--json"}, &out, &out, options{})
	if code != 0 {
		t.Fatalf("the measurement failed with the code %d:\n%s", code, out.String())
	}
	var report benchReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("the report is not JSON: %v\n%s", err, out.String())
	}
	if report.Commands != 1 {
		t.Errorf("the report says %d command(s), want 1", report.Commands)
	}
	if len(report.ByCommand) != 1 || report.ByCommand[0].Command != "say" {
		t.Errorf("the report has the commands %v, want one \u201csay\u201d", report.ByCommand)
	}
}

// TestBenchRepeatRunsAFreshWorld covers the repetition.
//
// Every repetition gets a fresh world, so a function that adds an
// objective does not fail on the second run.
func TestBenchRepeatRunsAFreshWorld(t *testing.T) {
	root := writeBenchFunction(t, "pack", "main",
		"scoreboard objectives add counter dummy\n")

	var out bytes.Buffer
	code := runBench([]string{root, "--all", "--repeat", "3", "--json"}, &out, &out, options{})
	if code != 0 {
		t.Fatalf("the measurement failed with the code %d:\n%s", code, out.String())
	}
	var report benchReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("the report is not JSON: %v", err)
	}
	if report.Failed != 0 {
		t.Errorf("the report says %d failure(s), want none", report.Failed)
	}
	if report.Commands != 3 {
		t.Errorf("the report says %d command(s), want 3", report.Commands)
	}
}

// TestBenchReportsAFunction covers the single function mode.
func TestBenchReportsAFunction(t *testing.T) {
	root := writeBenchFunction(t, "pack", "main", "say hello\n")

	var out bytes.Buffer
	code := runBench([]string{root, "--function", "pack:main"}, &out, &out, options{})
	if code != 0 {
		t.Fatalf("the measurement failed with the code %d:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "mode:   function pack:main") {
		t.Errorf("the report does not name the function:\n%s", out.String())
	}
}

// TestBenchWithoutAWorldFails covers the mistake of leaving the world out.
func TestBenchWithoutAWorldFails(t *testing.T) {
	var out bytes.Buffer
	if code := runBench(nil, &out, &out, options{}); code != 2 {
		t.Errorf("the code is %d, want 2", code)
	}
}
