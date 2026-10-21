package v1_12_2

import (
	"strings"
	"testing"
)

// TestParticleViewerIsAPlayer pins down the target variant of the viewer of
// `/particle`.
//
// 1.12.2 resolves the viewer with
// `CommandParticle.b(minecraftServer, listener, stringArray[10])`, which is
// `CommandAbstract.b` -- the **player** variant. It ends in
// `PlayerSelector.getPlayer` and yields an `EntityPlayer`, so a viewer
// that is not a player is refused.
//
// (Measured on a real 1.12.2 server, probe case 55-particle-viewer: with an
// armour stand in the world, `particle flame 0 4 0 0 0 0 0 1 normal
// @e[type=armor_stand]` is answered with "Player '<the uuid of the armour
// stand>' cannot be found" and the particle is not played.)
//
// The SPEC used to list the viewer among the commands that take an
// *entity*, which is what gomc implemented.
func TestParticleViewerIsAPlayer(t *testing.T) {
	e := newTestEngine(t)
	// An armour stand is not a player.
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}

	err := runLine(e, "particle flame 0 4 0 0 0 0 0 1 normal @e[type=armor_stand]")
	if err == nil {
		t.Fatal("a viewer that is not a player should be refused")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "Player '") || !strings.HasSuffix(msg, "' cannot be found") {
		t.Errorf("the error is %q, want the player wording", msg)
	}
	// The name is the uuid of the armour stand, not the selector text.
	if strings.Contains(msg, "@e[type=armor_stand]") {
		t.Errorf("the error names the selector %q, want the uuid of the entity", msg)
	}
}

// TestParticleViewerOfAListThatMatchesNothing covers the list case.
//
// A list selector that matches nothing makes the command run zero times and
// say nothing at all -- the rule of SPEC section 3.10, which holds for the
// player variant too.
//
// (Measured on a real 1.12.2 server, probe case 55-particle-viewer:
// `particle flame 0 4 0 0 0 0 0 1 normal @e[type=zombie]` prints nothing
// when there is no zombie in the world.)
func TestParticleViewerOfAListThatMatchesNothing(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "particle flame 0 4 0 0 0 0 0 1 normal @e[type=zombie]")
	if err != nil {
		t.Fatalf("particle: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("the command printed %q, want nothing", got)
	}
}

// TestParticleViewerOfAPlayer covers the case that does work: a viewer that
// is a player.
//
// (Measured on a real 1.12.2 server, probe case 55-particle-viewer:
// `particle flame 0 4 0 0 0 0 0 1 normal @p` plays the particle when a
// player is online.)
func TestParticleViewerOfAPlayer(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	got, err := runLineFeedback(e, "particle flame 0 4 0 0 0 0 0 1 normal @p")
	if err != nil {
		t.Fatalf("particle: %v", err)
	}
	if len(got) != 1 || got[0] != "Playing effect flame for 1 times" {
		t.Errorf("the command printed %q, want the success message", got)
	}
}

// TestParticleParameterRange pins down the numbers of the parameters the
// particle takes.
//
// 1.12.2 reads them with `Integer.parseInt`, so a value outside the signed
// 32-bit range is not a number at all. The parameters the particle does not
// take are **ignored**, not refused.
//
// (Measured on a real 1.12.2 server, probe case 55-particle-viewer:
// `particle blockcrack 0 4 0 0 0 0 0 1 normal Steve 9999999999` is answered
// with "Invalid param given (9999999999)", and the same command with an
// extra parameter behind the first one plays the particle without a
// complaint.)
func TestParticleParameterRange(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")

	for _, tc := range []struct {
		line string
		want string
	}{
		{"particle blockcrack 0 4 0 0 0 0 0 1 normal Steve 5", ""},
		{"particle blockcrack 0 4 0 0 0 0 0 1 normal Steve 5 99", ""},
		{"particle blockcrack 0 4 0 0 0 0 0 1 normal Steve notanumber", "Invalid param given (notanumber)"},
		{"particle blockcrack 0 4 0 0 0 0 0 1 normal Steve 9999999999", "Invalid param given (9999999999)"},
		{"particle blockcrack 0 4 0 0 0 0 0 1 normal Steve 2147483648", "Invalid param given (2147483648)"},
	} {
		err := runLine(e, tc.line)
		if tc.want == "" {
			if err != nil {
				t.Errorf("%q said %v, want no error", tc.line, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%q did not fail, want %q", tc.line, tc.want)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%q said %q, want %q", tc.line, err.Error(), tc.want)
		}
	}
}
