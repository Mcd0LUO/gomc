package difftest

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGoldenCases runs every case under testdata and compares it with its
// golden file.
//
// The golden files lock the behaviour of the runtime, so a change that
// breaks a 1.12.2 rule shows up here as a failing test rather than as a
// surprise on a real server.
func TestGoldenCases(t *testing.T) {
	pattern := filepath.Join("..", "..", "testdata", "cases", "*.difftest")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no cases found under %s", pattern)
	}
	world := filepath.Join("..", "..", "testdata", "demo-world")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			c, err := ParseCase(string(data))
			if err != nil {
				t.Fatalf("ParseCase: %v", err)
			}
			golden, err := os.ReadFile(path + ".golden")
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			obs, err := Run(world, c)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			got := obs.Render(ScopeServer)
			diff := Diff(string(golden), got)
			if len(diff) != 0 {
				t.Errorf("the case does not match its golden file:\n%s", FormatDiff(diff))
			}
		})
	}
}
