package world

import (
	"testing"
)

func TestScoreboardObjectiveLifecycle(t *testing.T) {
	sb := NewScoreboard()
	if _, created := sb.AddObjective("kills", CriteriaDummy, ""); !created {
		t.Fatal("AddObjective did not report a new objective")
	}
	// Adding the same objective again is not an error and does not
	// replace it.
	if _, created := sb.AddObjective("kills", CriteriaDummy, ""); created {
		t.Error("AddObjective reported a new objective for an existing one")
	}
	if sb.Objective("kills") == nil {
		t.Error("the objective is missing")
	}
	if !sb.RemoveObjective("kills") {
		t.Error("RemoveObjective did not remove the objective")
	}
	if sb.Objective("kills") != nil {
		t.Error("the objective survived the removal")
	}
}

func TestScoreboardSetAndReset(t *testing.T) {
	sb := NewScoreboard()
	sb.AddObjective("a", CriteriaDummy, "")

	if created := sb.SetScore("Steve", "a", 5); !created {
		t.Error("SetScore did not report a new holder")
	}
	if created := sb.SetScore("Steve", "a", 7); created {
		t.Error("SetScore reported a new holder for an existing one")
	}
	got, ok := sb.Score("Steve", "a")
	if !ok || got != 7 {
		t.Errorf("score = %d (ok=%v), want 7", got, ok)
	}

	sb.ResetScore("Steve", "a")
	if _, ok := sb.Score("Steve", "a"); ok {
		t.Error("the score survived the reset")
	}
}

// TestScoreboardArithmeticUsesJavaIntegers pins down the integer semantics.
//
// 1.12.2 stores every score in a Java `int`, so the results wrap around and
// the division truncates toward zero.
func TestScoreboardArithmeticUsesJavaIntegers(t *testing.T) {
	cases := []struct {
		op      Operation
		target  int
		operand int
		want    int
	}{
		{OpAdd, 2, 3, 5},
		{OpSub, 2, 3, -1},
		{OpMul, 4, 3, 12},
		{OpDiv, 7, 2, 3},
		{OpDiv, -7, 2, -3},
		{OpMod, 7, 3, 1},
		{OpMod, -7, 3, -1},
		{OpSet, 2, 3, 3},
		{OpKeepMin, 5, 9, 5},
		{OpKeepMax, 5, 9, 9},
		{OpSwap, 5, 9, 9},
		// The wrap around of a signed 32-bit integer.
		{OpAdd, 2147483647, 1, -2147483648},
		{OpSub, -2147483648, 1, 2147483647},
		{OpMul, 46341, 46341, -2147479015},
	}
	for _, c := range cases {
		got, err := Apply(c.op, c.target, c.operand)
		if err != nil {
			t.Fatalf("Apply(%q, %d, %d): %v", c.op, c.target, c.operand, err)
		}
		if got != c.want {
			t.Errorf("Apply(%q, %d, %d) = %d, want %d", c.op, c.target, c.operand, got, c.want)
		}
	}
}

func TestScoreboardArithmeticRejectsZeroDivisor(t *testing.T) {
	if _, err := Apply(OpDiv, 1, 0); err == nil {
		t.Error("expected an error for a division by zero")
	}
	if _, err := Apply(OpMod, 1, 0); err == nil {
		t.Error("expected an error for a modulo by zero")
	}
}

func TestScoreboardTeamMembership(t *testing.T) {
	sb := NewScoreboard()
	sb.AddTeam("red")
	sb.AddTeam("blue")

	if !sb.AddTeamMember("red", "Steve") {
		t.Fatal("AddTeamMember failed")
	}
	if got := sb.TeamNameOf("Steve"); got != "red" {
		t.Errorf("team = %q, want red", got)
	}

	// A holder belongs to at most one team.
	sb.AddTeamMember("blue", "Steve")
	if got := sb.TeamNameOf("Steve"); got != "blue" {
		t.Errorf("team = %q, want blue", got)
	}
	if t2 := sb.Team("red"); t2.Members["Steve"] {
		t.Error("Steve is still a member of the red team")
	}

	// An entity on no team reports an empty name.
	if got := sb.TeamNameOf("Alex"); got != "" {
		t.Errorf("team = %q, want an empty name", got)
	}

	sb.RemoveTeamMember("blue", "Steve")
	if got := sb.TeamNameOf("Steve"); got != "" {
		t.Errorf("team = %q, want an empty name", got)
	}
}

func TestGameruleDefaults(t *testing.T) {
	r := NewGameRules()
	if got := r.Int("maxCommandChainLength"); got != 65536 {
		t.Errorf("maxCommandChainLength = %d, want 65536", got)
	}
	if got := r.String("gameLoopFunction"); got != "-" {
		t.Errorf("gameLoopFunction = %q, want -", got)
	}
	if !r.Bool("commandBlockOutput") {
		t.Error("commandBlockOutput should default to true")
	}
	if r.Has("nosuchrule") {
		t.Error("Has reported an unknown rule")
	}
}

func TestGameruleSetRejectsBadValues(t *testing.T) {
	r := NewGameRules()
	if err := r.Set("commandBlockOutput", "maybe"); err == nil {
		t.Error("expected an error for a non-boolean value")
	}
	if err := r.Set("maxCommandChainLength", "many"); err == nil {
		t.Error("expected an error for a non-numeric value")
	}
	if err := r.Set("nosuchrule", "1"); err == nil {
		t.Error("expected an error for an unknown rule")
	}
	if err := r.Set("maxCommandChainLength", "10"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := r.Int("maxCommandChainLength"); got != 10 {
		t.Errorf("maxCommandChainLength = %d, want 10", got)
	}
}

// TestEntityBoundingBox pins down the collision box.
//
// 1.12.2 centres the box on the entity's X and Z and extends it upwards
// from the entity's feet.
func TestEntityBoundingBox(t *testing.T) {
	e := &Entity{Pos: Vec{X: 10, Y: 20, Z: 30}, Width: 0.6, Height: 1.95}
	box := e.BoundingBox()
	if box.MinX != 9.7 || box.MaxX != 10.3 {
		t.Errorf("x range = [%v, %v], want [9.7, 10.3]", box.MinX, box.MaxX)
	}
	if box.MinY != 20 || box.MaxY != 21.95 {
		t.Errorf("y range = [%v, %v], want [20, 21.95]", box.MinY, box.MaxY)
	}
	if box.MinZ != 29.7 || box.MaxZ != 30.3 {
		t.Errorf("z range = [%v, %v], want [29.7, 30.3]", box.MinZ, box.MaxZ)
	}
}

func TestAABBIntersects(t *testing.T) {
	a := AABB{MinX: 0, MinY: 0, MinZ: 0, MaxX: 1, MaxY: 1, MaxZ: 1}
	if !a.Intersects(AABB{MinX: 0.5, MinY: 0.5, MinZ: 0.5, MaxX: 2, MaxY: 2, MaxZ: 2}) {
		t.Error("the boxes overlap but Intersects said otherwise")
	}
	if a.Intersects(AABB{MinX: 2, MinY: 2, MinZ: 2, MaxX: 3, MaxY: 3, MaxZ: 3}) {
		t.Error("the boxes are apart but Intersects said they overlap")
	}
}
