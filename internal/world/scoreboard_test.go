package world

import (
	"fmt"
	"reflect"
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

// TestScoreboardRemoveHolderFromTeam pins down the leaving of a team.
//
// 1.12.2's `Scoreboard.removePlayerFromTeam(String)` takes the player
// out of the team it is on and reports whether it was on a team at all,
// so the command can tell the two cases apart.
func TestScoreboardRemoveHolderFromTeam(t *testing.T) {
	sb := NewScoreboard()
	sb.AddTeam("red")
	sb.AddTeam("blue")
	sb.AddTeamMember("red", "Steve")
	sb.AddTeamMember("red", "Alex")

	if !sb.RemoveHolderFromTeam("Steve") {
		t.Error("RemoveHolderFromTeam did not report the removal")
	}
	if got := sb.TeamNameOf("Steve"); got != "" {
		t.Errorf("team = %q, want an empty name", got)
	}
	if got := len(sb.Team("red").Members); got != 1 {
		t.Errorf("the red team has %d members, want 1", got)
	}

	// A holder that is on no team cannot be removed from one.
	if sb.RemoveHolderFromTeam("Steve") {
		t.Error("RemoveHolderFromTeam reported a removal for a holder on no team")
	}
}

// TestScoreboardTrackedPlayers pins down the list of tracked players.
//
// 1.12.2's `Scoreboard.getPlayers()` returns the keys of the map that
// holds the scores, so a player is listed once no matter how many
// objectives it has a score in.
func TestScoreboardTrackedPlayers(t *testing.T) {
	sb := NewScoreboard()
	sb.AddObjective("kills", CriteriaDummy, "")
	sb.AddObjective("deaths", CriteriaDummy, "")

	if got := sb.TrackedPlayers(); len(got) != 0 {
		t.Errorf("tracked players = %v, want none", got)
	}

	sb.SetScore("Steve", "kills", 3)
	sb.SetScore("Steve", "deaths", 1)
	sb.SetScore("Alex", "kills", 2)

	got := sb.TrackedPlayers()
	if len(got) != 2 || got[0] != "Alex" || got[1] != "Steve" {
		t.Errorf("tracked players = %v, want [Alex Steve]", got)
	}
}

// TestScoreboardTeamMemberList pins down the member list of a team.
func TestScoreboardTeamMemberList(t *testing.T) {
	sb := NewScoreboard()
	sb.AddTeam("red")
	sb.AddTeamMember("red", "Steve")
	sb.AddTeamMember("red", "Alex")

	got := sb.Team("red").MemberList()
	if len(got) != 2 || got[0] != "Alex" || got[1] != "Steve" {
		t.Errorf("members = %v, want [Alex Steve]", got)
	}
}

// objectiveLabel returns the name 1.12.2 shows for an objective.
func objectiveLabel(o *Objective) string {
	if o.DisplayName == "" {
		return o.Name
	}
	return o.DisplayName
}

// scannedTrackedObjectives reproduces the way `scoreboard players list`
// found the tracked objectives of a holder before the scoreboard kept
// an index of them: it walked every objective and asked for the score
// of the holder.
//
// The lines are the ones `scoreboard players list` prints, so the two
// helpers can be compared line by line.
func scannedTrackedObjectives(sb *Scoreboard, holder string) []string {
	var out []string
	for _, o := range sb.Objectives() {
		if v, ok := sb.Score(holder, o.Name); ok {
			out = append(out, fmt.Sprintf("- %s: %d (%s)", objectiveLabel(o), v, o.Name))
		}
	}
	return out
}

// indexedTrackedObjectives renders the tracked objectives of a holder
// the way `scoreboard players list` does, but through the index of the
// scoreboard.
func indexedTrackedObjectives(sb *Scoreboard, holder string) []string {
	var out []string
	for _, t := range sb.TrackedObjectives(holder) {
		out = append(out, fmt.Sprintf("- %s: %d (%s)", objectiveLabel(t.Objective), t.Score(holder), t.Objective.Name))
	}
	return out
}

// TestScoreboardTrackedObjectivesMatchTheScan is the regression test of
// the index of the tracked objectives.
//
// The index must answer exactly what the scan over every objective
// answered: the same objectives, the same order and the same scores.
// The case covers a scoreboard with many objectives and many holders,
// because that is where the scan was quadratic.
func TestScoreboardTrackedObjectivesMatchTheScan(t *testing.T) {
	const (
		objectives = 64
		holders    = 50
	)
	sb := NewScoreboard()
	for i := 0; i < objectives; i++ {
		sb.AddObjective(fmt.Sprintf("obj%d", i), CriteriaDummy, "")
	}
	// Every third holder has a score in every third objective, so the
	// holders are tracked in different objectives.
	for i := 0; i < objectives; i++ {
		name := fmt.Sprintf("obj%d", i)
		for h := 0; h < holders; h++ {
			if (h+i)%3 == 0 {
				sb.SetScore(fmt.Sprintf("holder%d", h), name, h+i)
			}
		}
	}

	check := func(what string) {
		t.Helper()
		for h := 0; h < holders; h++ {
			holder := fmt.Sprintf("holder%d", h)
			want := scannedTrackedObjectives(sb, holder)
			got := indexedTrackedObjectives(sb, holder)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s: the listing of %s is %v, want %v", what, holder, got, want)
			}
			if len(want) != len(sb.TrackedObjectives(holder)) {
				t.Fatalf("%s: %s is tracked in %d objectives, want %d",
					what, holder, len(sb.TrackedObjectives(holder)), len(want))
			}
		}
	}

	check("after the scores were set")

	// The index has to follow the scoreboard when it changes.
	sb.ResetScore("holder1", "obj1")
	sb.ResetScore("holder2", "")
	check("after the scores were reset")

	sb.RemoveObjective("obj3")
	check("after an objective was removed")

	sb.AddObjective("obj3", CriteriaDummy, "")
	sb.SetScore("holder1", "obj3", 7)
	check("after the objective was added again")

	sb.SetScore("holder4", "obj5", 9)
	check("after a score was set")
}

// TestScoreboardTrackedObjectivesKeepTheCreationOrder pins down the
// order of the listing.
//
// 1.12.2 lists the objectives of a player in the order the objectives
// were created, not in the order the scores were set, and the listing
// of a player with many objectives has one line per objective.
func TestScoreboardTrackedObjectivesKeepTheCreationOrder(t *testing.T) {
	const objectives = 40
	sb := NewScoreboard()
	for i := 0; i < objectives; i++ {
		sb.AddObjective(fmt.Sprintf("obj%d", i), CriteriaDummy, "")
	}
	// The scores are set backwards, so the order of the scores says
	// nothing about the order of the listing.
	for i := objectives - 1; i >= 0; i-- {
		sb.SetScore("Steve", fmt.Sprintf("obj%d", i), i)
	}

	tracked := sb.TrackedObjectives("Steve")
	if len(tracked) != objectives {
		t.Fatalf("Steve is tracked in %d objectives, want %d", len(tracked), objectives)
	}
	for i, tracked := range tracked {
		want := fmt.Sprintf("obj%d", i)
		if tracked.Objective.Name != want {
			t.Fatalf("objective %d is %q, want %q", i, tracked.Objective.Name, want)
		}
		if got := tracked.Score("Steve"); got != i {
			t.Fatalf("the score of %s in %s is %d, want %d", "Steve", want, got, i)
		}
	}

	// The lines are the ones `scoreboard players list` prints.
	want := []string{"- obj0: 0 (obj0)", "- obj1: 1 (obj1)", "- obj2: 2 (obj2)"}
	got := indexedTrackedObjectives(sb, "Steve")
	if len(got) != objectives {
		t.Fatalf("the listing has %d lines, want %d", len(got), objectives)
	}
	for i, line := range want {
		if got[i] != line {
			t.Errorf("line %d is %q, want %q", i, got[i], line)
		}
	}
}

// TestScoreboardTrackedObjectivesNeedsAScore pins down the difference
// between an objective that exists and a holder that is tracked in it.
//
// 1.12.2 answers "Player %s has no scores recorded" for a player that
// is not tracked in any objective, even when the scoreboard has
// objectives.
func TestScoreboardTrackedObjectivesNeedsAScore(t *testing.T) {
	sb := NewScoreboard()
	sb.AddObjective("kills", CriteriaDummy, "")
	sb.SetScore("Alex", "kills", 1)

	if got := sb.TrackedObjectives("Steve"); len(got) != 0 {
		t.Errorf("Steve is tracked in %v, want nothing", got)
	}
	if got := sb.TrackedObjectives("Alex"); len(got) != 1 {
		t.Errorf("Alex is tracked in %d objectives, want 1", len(got))
	}
}

// TestScoreboardTrackedObjectivesIgnoresUnknownObjectives pins down the
// scores of an objective that does not exist.
//
// 1.12.2 does not list the scores of an objective that is not there,
// and it lists them once the objective is added.
func TestScoreboardTrackedObjectivesIgnoresUnknownObjectives(t *testing.T) {
	sb := NewScoreboard()
	sb.AddObjective("kills", CriteriaDummy, "")
	sb.SetScore("Steve", "nosuch", 3)

	if got := sb.TrackedObjectives("Steve"); len(got) != 0 {
		t.Errorf("Steve is tracked in %v, want nothing", got)
	}

	sb.AddObjective("nosuch", CriteriaDummy, "")
	got := sb.TrackedObjectives("Steve")
	if len(got) != 1 || got[0].Objective.Name != "nosuch" {
		t.Fatalf("Steve is tracked in %v, want [nosuch]", got)
	}
	if v := got[0].Score("Steve"); v != 3 {
		t.Errorf("the score of Steve in nosuch is %d, want 3", v)
	}
}

// TestScoreboardTrackedObjectivesFollowTheScore pins down that the
// listing shows the current score of a holder.
func TestScoreboardTrackedObjectivesFollowTheScore(t *testing.T) {
	sb := NewScoreboard()
	sb.AddObjective("kills", CriteriaDummy, "")
	sb.SetScore("Steve", "kills", 3)

	if v := sb.TrackedObjectives("Steve")[0].Score("Steve"); v != 3 {
		t.Fatalf("the score of Steve is %d, want 3", v)
	}
	sb.SetScore("Steve", "kills", 9)
	if v := sb.TrackedObjectives("Steve")[0].Score("Steve"); v != 9 {
		t.Errorf("the score of Steve is %d, want 9", v)
	}
}

// TestScoreboardAddTeamMemberNeedsATeam pins down the joining of a team.
//
// 1.12.2's `Scoreboard.addPlayerToTeam` answers false when the team
// does not exist, and that is the only way the join command can fail.
func TestScoreboardAddTeamMemberNeedsATeam(t *testing.T) {
	sb := NewScoreboard()
	if sb.AddTeamMember("nosuchteam", "Steve") {
		t.Error("AddTeamMember reported a join into a team that does not exist")
	}
	sb.AddTeam("red")
	if !sb.AddTeamMember("red", "Steve") {
		t.Error("AddTeamMember did not report the join")
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
	// 1.12.2 creates a rule that does not exist yet: `GameRules.set`
	// falls back to ANY_VALUE and adds it.
	if err := r.Set("nosuchrule", "1"); err != nil {
		t.Errorf("setting an unknown rule: %v", err)
	}
	if got := r.String("nosuchrule"); got != "1" {
		t.Errorf("nosuchrule = %q, want %q", got, "1")
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
