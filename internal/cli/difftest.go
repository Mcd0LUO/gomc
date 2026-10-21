package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/difftest"
)

// runDifftest dispatches the difftest subcommands.
func runDifftest(args []string, stdout, stderr io.Writer, opts options) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "difftest requires a subcommand: run, check, update, probe or list")
		return 2
	}
	if wantsHelp(args) {
		usageDifftest(stdout)
		return 0
	}
	switch args[0] {
	case "run":
		return difftestRun(args[1:], stdout, stderr)
	case "check":
		return difftestCheck(args[1:], stdout, stderr)
	case "update":
		return difftestUpdate(args[1:], stdout, stderr)
	case "probe":
		return difftestProbe(args[1:], stdout, stderr)
	case "list":
		return difftestList(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "unknown difftest subcommand %q\n", args[0])
	return 2
}

// difftestList lists the cases of a directory and their golden files.
//
// The cases are the files of the repository, so the directory defaults to
// "testdata/cases" — the place they live in.
func difftestList(args []string, stdout, stderr io.Writer) int {
	dir := "testdata/cases"
	if len(args) == 1 {
		dir = args[0]
	} else if len(args) > 1 {
		fmt.Fprintln(stderr, "list takes at most one directory")
		return 2
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot read %s: %v\n", dir, err)
		return 1
	}
	count := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".difftest") {
			continue
		}
		path := filepath.Join(dir, name)
		c, err := readCase(path)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		golden := "no golden"
		if _, err := os.Stat(goldenPath(path)); err == nil {
			golden = "golden"
		}
		fmt.Fprintf(stdout, "%-34s %3d command(s)  %s\n", c.Name, len(c.Run), golden)
		count++
	}
	fmt.Fprintf(stdout, "\n%d case(s) in %s\n", count, dir)
	return 0
}

// readCase loads a case file.
func readCase(path string) (*difftest.Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := difftest.ParseCase(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

// goldenPath returns the golden file path of a case.
func goldenPath(casePath string) string {
	return casePath + ".golden"
}

// parseDifftestArgs splits the world and the case path out of the
// arguments, and reports whether the full state was requested.
func parseDifftestArgs(args []string) (world, casePath string, full bool, err error) {
	for _, a := range args {
		switch {
		case a == "--full":
			full = true
		case strings.HasPrefix(a, "-"):
			return "", "", false, fmt.Errorf("unknown option %q", a)
		case world == "":
			world = a
		case casePath == "":
			casePath = a
		default:
			return "", "", false, fmt.Errorf("too many arguments")
		}
	}
	if world == "" || casePath == "" {
		return "", "", false, fmt.Errorf("a world directory and a case file are required")
	}
	return world, casePath, full, nil
}

// scopeOf picks the observation scope.
func scopeOf(full bool) difftest.Scope {
	if full {
		return difftest.ScopeFull
	}
	return difftest.ScopeServer
}

// difftestRun prints the observation of a case.
func difftestRun(args []string, stdout, stderr io.Writer) int {
	world, casePath, full, err := parseDifftestArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	c, err := readCase(casePath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	obs, err := difftest.Run(world, c)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, obs.Render(scopeOf(full)))
	return 0
}

// difftestCheck compares a case with its golden file.
func difftestCheck(args []string, stdout, stderr io.Writer) int {
	world, casePath, full, err := parseDifftestArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	c, err := readCase(casePath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	golden := goldenPath(casePath)
	want, err := os.ReadFile(golden)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot read the golden file %s: %v\n", golden, err)
		fmt.Fprintf(stderr, "hint: run `gomc difftest update %s %s` to create it\n", world, casePath)
		return 1
	}
	obs, err := difftest.Run(world, c)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	got := obs.Render(scopeOf(full))
	diff := difftest.Diff(string(want), got)
	if len(diff) == 0 {
		fmt.Fprintf(stdout, "ok: %s matches %s\n", casePath, filepath.Base(golden))
		return 0
	}
	fmt.Fprintf(stderr, "error: %s does not match %s\n", casePath, filepath.Base(golden))
	for _, line := range diff {
		fmt.Fprintln(stderr, line)
	}
	return 1
}

// difftestUpdate rewrites the golden file of a case.
func difftestUpdate(args []string, stdout, stderr io.Writer) int {
	world, casePath, full, err := parseDifftestArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	c, err := readCase(casePath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	obs, err := difftest.Run(world, c)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	golden := goldenPath(casePath)
	if err := os.WriteFile(golden, []byte(obs.Render(scopeOf(full))), 0o644); err != nil {
		fmt.Fprintf(stderr, "error: cannot write %s: %v\n", golden, err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", golden)
	return 0
}

// difftestProbe writes the probe function for a real server.
func difftestProbe(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "probe requires a case file")
		return 2
	}
	c, err := readCase(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, difftest.Probe(c))
	return 0
}
