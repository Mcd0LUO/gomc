package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// locateEngine builds an engine with the 1.12.2 command set.
//
// The locate command is registered on its own: `Register` does not add it
// yet, because the file that does is written by another hand.
func locateEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	Register(reg)
	registerLocate(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// locateError runs a command line that has to fail and returns the message.
func locateError(t *testing.T, e *exec.Engine, line string) string {
	t.Helper()
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	_, err := e.ExecuteLine(line, 1, src)
	if err == nil {
		t.Fatalf("%q: expected an error", line)
	}
	return err.Error()
}

// locateErrorOf runs a command line that has to fail and returns the error.
func locateErrorOf(t *testing.T, e *exec.Engine, line string) *exec.Error {
	t.Helper()
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	_, err := e.ExecuteLine(line, 1, src)
	if err == nil {
		t.Fatalf("%q: expected an error", line)
	}
	ee, ok := err.(*exec.Error)
	if !ok {
		t.Fatalf("%q: the error is a %T, want an *exec.Error", line, err)
	}
	return ee
}

// TestLocateUsage checks the wording of a usage mistake.
//
// 1.12.2 reads exactly one argument, so both a missing argument and an
// extra one are answered with the usage of the command.
//
// (Measured on a real 1.12.2 server.)
func TestLocateUsage(t *testing.T) {
	e := locateEngine(t)

	want := "Usage: /locate <feature>"
	for _, line := range []string{
		"locate",
		"locate biome plains",
		"locate structure village",
		"locate plains plains",
		"locate village extra",
	} {
		if got := locateError(t, e, line); got != want {
			t.Errorf("%q: the error was %q, want %q", line, got, want)
		}
	}
}

// TestLocateUnknownFeature checks the wording of a feature the server does
// not know.
//
// The server does not tell a name it does not know from a structure it
// cannot reach: both are answered the same way.
//
// (Measured on a real 1.12.2 server: `locate plains` and `locate bogus` are
// both answered with "Unable to locate any <name> feature".)
func TestLocateUnknownFeature(t *testing.T) {
	e := locateEngine(t)

	for _, tc := range []struct {
		line string
		want string
	}{
		{"locate plains", "Unable to locate any plains feature"},
		{"locate bogus", "Unable to locate any bogus feature"},
		{"locate biome", "Unable to locate any biome feature"},
		{"locate structure", "Unable to locate any structure feature"},
	} {
		if got := locateError(t, e, tc.line); got != tc.want {
			t.Errorf("%q: the error was %q, want %q", tc.line, got, tc.want)
		}
	}
}

// TestLocateKeepsTheName checks that the name is repeated as it was
// written.
//
// 1.12.2 puts the argument into the message unchanged, so the case of the
// name is part of the answer.
//
// (Measured on a real 1.12.2 server: `locate stronghold` is answered with
// "Unable to locate any stronghold feature".)
func TestLocateKeepsTheName(t *testing.T) {
	e := locateEngine(t)

	want := "Unable to locate any stronghold feature"
	if got := locateError(t, e, "locate stronghold"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	want = "Unable to locate any VILLAGE feature"
	if got := locateError(t, e, "locate VILLAGE"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestLocateStructures checks the wording for the structures the server
// can find in the overworld.
//
// gomc does not simulate world generation, so it never finds one. The
// wording is the server's own, and the hint says why gomc stops there.
//
// (Measured on a real 1.12.2 server: `locate Stronghold` there is answered
// with "Located Stronghold at 808 (y?) -1144".)
func TestLocateStructures(t *testing.T) {
	e := locateEngine(t)

	for _, name := range []string{
		"Stronghold", "Monument", "Village", "Mansion", "Temple", "Mineshaft",
	} {
		line := "locate " + name
		want := "Unable to locate any " + name + " feature"
		if got := locateError(t, e, line); got != want {
			t.Errorf("%q: the error was %q, want %q", line, got, want)
		}
		if hint := locateErrorOf(t, e, line).Hint; hint == "" {
			t.Errorf("%q: the error has no hint", line)
		}
	}
}

// TestLocateOtherDimension checks the structures that are not in the
// overworld.
//
// EndCity belongs to the End and Fortress to the nether, so the server
// cannot find them in the overworld either -- and says the same thing gomc
// says.
//
// (Measured on a real 1.12.2 server: `locate EndCity` and `locate Fortress`
// are answered with "Unable to locate any <name> feature".)
func TestLocateOtherDimension(t *testing.T) {
	e := locateEngine(t)

	for _, name := range []string{"EndCity", "Fortress"} {
		line := "locate " + name
		want := "Unable to locate any " + name + " feature"
		if got := locateError(t, e, line); got != want {
			t.Errorf("%q: the error was %q, want %q", line, got, want)
		}
		if hint := locateErrorOf(t, e, line).Hint; hint != "" {
			t.Errorf("%q: the error has the hint %q, want none", line, hint)
		}
	}
}

// TestLocatePermission checks the permission level of the command.
//
// 1.12.2 asks the command which level it needs (`CommandLocate.a`), and it
// answers 2.
func TestLocatePermission(t *testing.T) {
	e := locateEngine(t)

	src := &exec.Source{Name: "Steve", World: "world", PermissionLevel: 1, Feedback: true}
	_, err := e.ExecuteLine("locate Stronghold", 1, src)
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "You do not have permission to use this command"
	if got := err.Error(); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}
