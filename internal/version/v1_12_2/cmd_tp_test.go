package v1_12_2

import (
	"math"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// TestTpToAPlayer covers the second form of `/tp`.
//
// 1.12.2 decides the form of the command by the number of arguments: one
// or two of them mean "to an entity". The destination is the last
// argument, and the target is the first one.
//
// (Measured on a real 1.12.2 server: `tp @s @p[tag=target]` teleports to
// the player, and the message is "Teleported <a> to <b>" without the
// coordinates.)
func TestTpToAPlayer(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err := runLineFeedback(e, "tp @e[type=armor_stand] Steve")
	if err != nil {
		t.Fatalf("tp: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "Teleported Armor Stand to Steve" {
		t.Errorf("feedback = %v, want the teleport to the player", feedback)
	}
	for _, ent := range e.World.EntityList() {
		if ent.Type == "armor_stand" && ent.Pos != (world.Vec{X: 1, Y: 4, Z: 1}) {
			t.Errorf("the armor stand is at %v, want the position of the player", ent.Pos)
		}
	}
}

// TestTpWithoutATarget covers the form of `/tp` that has no target.
//
// 1.12.2 teleports the player who runs the command when the first
// argument is not a target, which is the case when there are three
// arguments (the coordinates).
func TestTpWithoutATarget(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	src := &exec.Source{
		Name: "Steve", World: "world", PermissionLevel: 4, Feedback: true,
		Entity: e.World.FindPlayer("Steve"), Pos: world.Vec{X: 1, Y: 4, Z: 1},
		ListenerPlayer: true,
	}
	if _, err := e.ExecuteLine("tp 5 6 7", 1, src); err != nil {
		t.Fatalf("tp: %v", err)
	}
	p := e.World.FindPlayer("Steve")
	if p == nil {
		t.Fatal("the player is gone")
	}
	// The X and the Z are in the "centre" mode: a whole number names
	// the middle of a block, so the player ends up at 5.5 and 7.5.
	if p.Pos != (world.Vec{X: 5.5, Y: 6, Z: 7.5}) {
		t.Errorf("the player is at %v, want (5.5, 6, 7.5)", p.Pos)
	}
}

// TestTPRotation pins down the yaw and the pitch of a `/tp`.
//
// 1.12.2 reads them the way it reads a coordinate: a number is an
// absolute angle, a number behind a tilde is measured from the angle the
// entity has, and a bare tilde keeps it. The yaw is folded into
// [-180, 180) and the pitch is folded first and clamped to [-90, 90]
// after that.
//
// (Measured on a real 1.12.2 server, probe cases 59-tp-rotation and
// 60-tp-rotation-bounds.)
func TestTPRotation(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	stand := e.World.EntitySlice()[0]

	// An absolute yaw and pitch.
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 90 45"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Yaw != 90 {
		t.Errorf("the yaw is %v, want 90", stand.Rot.Yaw)
	}
	if stand.Rot.Pitch != 45 {
		t.Errorf("the pitch is %v, want 45", stand.Rot.Pitch)
	}

	// A tilde is measured from the angle the entity has. The yaw of
	// 180 folds to -180, because the fold takes an angle of 180 to the
	// lower half (`MathHelper.g`).
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 ~90 0"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Yaw != -180 {
		t.Errorf("the yaw is %v, want -180", stand.Rot.Yaw)
	}

	// The yaw is folded into [-180, 180).
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 270 0"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Yaw != -90 {
		t.Errorf("the yaw is %v, want -90", stand.Rot.Yaw)
	}
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 400 0"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Yaw != 40 {
		t.Errorf("the yaw is %v, want 40", stand.Rot.Yaw)
	}

	// The pitch is folded first and clamped after that, so a pitch of
	// 200 ends up at -90 and not at 90.
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 0 200"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Pitch != -90 {
		t.Errorf("the pitch is %v, want -90", stand.Rot.Pitch)
	}
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 0 -200"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Pitch != 90 {
		t.Errorf("the pitch is %v, want 90", stand.Rot.Pitch)
	}
}

// TestTPRotationKeepsTheAngleItWasNotGiven covers the angle the command
// does not name.
//
// 1.12.2 reads the yaw and the pitch with the same default: a bare tilde,
// which keeps the angle the entity has.
func TestTPRotationKeepsTheAngleItWasNotGiven(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	stand := e.World.EntitySlice()[0]
	stand.Rot.Yaw = 33
	stand.Rot.Pitch = 12

	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 ~ ~"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	if stand.Rot.Yaw != 33 {
		t.Errorf("the yaw is %v, want 33", stand.Rot.Yaw)
	}
	if stand.Rot.Pitch != 12 {
		t.Errorf("the pitch is %v, want 12", stand.Rot.Pitch)
	}
}

// TestTPRotationIsFoundByTheSelector covers the point of the whole
// feature: an entity that was turned can be found by a selector that
// filters on the rotation.
//
// (This is what the real projects rely on: 74526 lines of the Wukong
// project select on `ry`/`rym`, and they only find an entity that has
// a rotation.)
func TestTPRotationIsFoundByTheSelector(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "tp @e[type=armor_stand] 1 4 1 90 45"); err != nil {
		t.Fatalf("tp: %v", err)
	}
	got, err := runLineFeedback(e, "execute @e[type=armor_stand,ry=90] ~ ~ ~ say hit")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("the selector found %d entities, want 1", len(got))
	}
}

// TestFoldAngle and TestClampAngle cover the two helpers.
func TestFoldAngle(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want float64
	}{
		{0, 0},
		{90, 90},
		{180, -180},
		{270, -90},
		{400, 40},
		{-90, -90},
		{-270, 90},
	} {
		if got := foldAngle(tc.in); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("foldAngle(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestClampAngle(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want float64
	}{
		{0, 0},
		{90, 90},
		{91, 90},
		{-91, -90},
		{1000, 90},
	} {
		if got := clampAngle(tc.in); got != tc.want {
			t.Errorf("clampAngle(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
