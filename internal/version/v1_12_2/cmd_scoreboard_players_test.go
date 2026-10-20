package v1_12_2

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// uuidPattern is the shape of a 1.12.2 UUID: eight, four, four, four and
// twelve lower case hex digits.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// addArmourStand puts an armour stand with the given tags into the world.
//
// The tags are the ones 1.12.2 keeps in the `Tags` key of the entity's
// NBT, so a dataTag that names them matches.
func addArmourStand(t *testing.T, e *exec.Engine, id string, tags ...string) *world.Entity {
	t.Helper()
	s, ok := e.World.EntityTypes.NewEntity("armor_stand", id)
	if !ok {
		t.Fatalf("the armor stand entity type is missing")
	}
	for _, tag := range tags {
		s.SetTag(tag, true)
	}
	e.World.AddEntity(s)
	return s
}

// TestEntityUUIDLooksLikeTheRealOne pins down the shape of the UUID the
// runtime gives an entity.
//
// 1.12.2 gives every entity a random UUID and prints it in the shape
// `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, in lower case.
//
// (Measured on a real 1.12.2 server: `scoreboard players set @e[type=zombie]
// kills 5` answers "Set score of kills for player
// 8ec2ce84-37a8-4713-8d8f-c72daff632a1 to 5".)
func TestEntityUUIDLooksLikeTheRealOne(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1")
	uuid := s.EnsureUUID()
	if !uuidPattern.MatchString(uuid) {
		t.Errorf("the UUID is %q, want the shape of a 1.12.2 UUID", uuid)
	}
	if uuid != strings.ToLower(uuid) {
		t.Errorf("the UUID is %q, want it in lower case", uuid)
	}
}

// TestEntityUUIDIsDeterministic pins down that the UUID of an entity does
// not change from run to run.
//
// The runtime is a deterministic interpreter, so the UUID cannot be drawn
// at random: it is derived from the entity's internal id. The same entity
// therefore always gets the same UUID, and two entities never get the
// same one.
func TestEntityUUIDIsDeterministic(t *testing.T) {
	first := newTestEngine(t)
	second := newTestEngine(t)
	a := addArmourStand(t, first, "a1")
	b := addArmourStand(t, first, "a2")
	c := addArmourStand(t, second, "a1")

	if a.EnsureUUID() != c.EnsureUUID() {
		t.Errorf("the UUID of the same entity changed: %q and %q",
			a.EnsureUUID(), c.EnsureUUID())
	}
	if a.EnsureUUID() == b.EnsureUUID() {
		t.Errorf("two entities share the UUID %q", a.EnsureUUID())
	}
	// Asking twice does not change the answer either.
	if a.EnsureUUID() != a.EnsureUUID() {
		t.Errorf("the UUID of an entity changed between two calls")
	}
}

// TestScoreboardHoldsAnEntityUnderItsUUID pins down that the scoreboard
// tracks an entity that is not a player under its UUID.
//
// (Measured on a real 1.12.2 server: the score of a zombie is kept under
// its UUID, and `scoreboard players list` prints that UUID.)
func TestScoreboardHoldsAnEntityUnderItsUUID(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1")
	uuid := s.EnsureUUID()

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	got, err := runLineFeedback(e, "scoreboard players set @e[type=armor_stand] kills 5")
	if err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	want := "Set score of kills for player " + uuid + " to 5"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("the command said %v, want %q", got, want)
	}
	// The score is kept under the UUID, not under the entity id.
	if _, tracked := e.World.Scoreboard.Score(uuid, "kills"); !tracked {
		t.Errorf("the score of %q is not tracked under its UUID", uuid)
	}
	if _, tracked := e.World.Scoreboard.Score("a1", "kills"); tracked {
		t.Errorf("the score is tracked under the entity id as well")
	}
}

// TestScoreboardPlayersSetByUUID pins down that the UUID of an entity is
// a score holder of its own.
//
// (Measured on a real 1.12.2 server: `scoreboard players set
// 8ec2ce84-37a8-4713-8d8f-c72daff632a1 kills 5` answers "Set score of
// kills for player 8ec2ce84-37a8-4713-8d8f-c72daff632a1 to 5".)
func TestScoreboardPlayersSetByUUID(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1")
	uuid := s.EnsureUUID()

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	got, err := runLineFeedback(e, "scoreboard players set "+uuid+" kills 5")
	if err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	want := "Set score of kills for player " + uuid + " to 5"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("the command said %v, want %q", got, want)
	}
}

// TestScoreboardPlayersTagByUUID pins down that the UUID of an entity
// finds the entity for a tag action.
//
// (Measured on a real 1.12.2 server: `scoreboard players tag
// 8ec2ce84-37a8-4713-8d8f-c72daff632a1 add marked` answers "Tag marked
// added".)
func TestScoreboardPlayersTagByUUID(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1")
	uuid := s.EnsureUUID()

	got, err := runLineFeedback(e, "scoreboard players tag "+uuid+" add marked")
	if err != nil {
		t.Fatalf("scoreboard players tag add: %v", err)
	}
	if len(got) != 1 || got[0] != "Tag marked added" {
		t.Fatalf("the command said %v, want %q", got, "Tag marked added")
	}
	if !s.HasTag("marked") {
		t.Error("the entity does not carry the tag")
	}
	// The listing names the entity by its UUID, the way the scoreboard
	// does.
	got, err = runLineFeedback(e, "scoreboard players tag "+uuid+" list")
	if err != nil {
		t.Fatalf("scoreboard players tag list: %v", err)
	}
	want := []string{"Tags on entity " + uuid + " are:", "marked"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("the listing was %v, want %v", got, want)
	}
}

// TestScoreboardPlayersDataTagMatches pins down that a dataTag that
// matches the entity lets the command through.
//
// (Measured on a real 1.12.2 server: `scoreboard players set
// @e[type=armor_stand] kills 2 {Tags:["mytag"]}` answers "Set score of
// kills for player <uuid> to 2".)
func TestScoreboardPlayersDataTagMatches(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1", "mytag")
	uuid := s.EnsureUUID()

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	got, err := runLineFeedback(e,
		`scoreboard players set @e[type=armor_stand] kills 2 {Tags:["mytag"]}`)
	if err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	want := "Set score of kills for player " + uuid + " to 2"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("the command said %v, want %q", got, want)
	}
	if score, tracked := e.World.Scoreboard.Score(uuid, "kills"); !tracked || score != 2 {
		t.Errorf("the score is %d (tracked: %v), want 2", score, tracked)
	}
}

// TestScoreboardPlayersDataTagMismatch pins down that a dataTag that does
// not match the entity stops the command and leaves the score alone.
//
// (Measured on a real 1.12.2 server: `scoreboard players set
// @e[type=armor_stand] kills 3 {Tags:["other"]}` answers "The dataTag
// does not match for <uuid>", and the score of the entity stays what it
// was.)
func TestScoreboardPlayersDataTagMismatch(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1", "mytag")
	uuid := s.EnsureUUID()

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set "+uuid+" kills 1"); err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 3 {Tags:["other"]}`)
	if err == nil {
		t.Fatal("expected an error for a dataTag that does not match")
	}
	want := "The dataTag does not match for " + uuid
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	// The score is the one of before: the check runs before the write.
	if score, _ := e.World.Scoreboard.Score(uuid, "kills"); score != 1 {
		t.Errorf("the score is %d, want the old one (1)", score)
	}
}

// TestScoreboardPlayersDataTagIsCheckedByAddAndRemove pins down that the
// dataTag is checked by `add` and `remove` as well.
//
// The language file only has a wording for `set`, but the three actions
// share one implementation in 1.12.2, so all three check the dataTag.
//
// (Measured on a real 1.12.2 server: `scoreboard players add
// @e[type=armor_stand] kills 1 {Tags:["other"]}` answers "The dataTag does
// not match for <uuid>".)
func TestScoreboardPlayersDataTagIsCheckedByAddAndRemove(t *testing.T) {
	for _, action := range []string{"add", "remove"} {
		e := newTestEngine(t)
		s := addArmourStand(t, e, "a1", "mytag")
		uuid := s.EnsureUUID()

		if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
			t.Fatalf("scoreboard objectives add: %v", err)
		}
		if err := runLine(e, "scoreboard players set "+uuid+" kills 5"); err != nil {
			t.Fatalf("scoreboard players set: %v", err)
		}
		err := runLine(e,
			`scoreboard players `+action+` @e[type=armor_stand] kills 1 {Tags:["other"]}`)
		if err == nil {
			t.Fatalf("%s: expected an error for a dataTag that does not match", action)
		}
		want := "The dataTag does not match for " + uuid
		if err.Error() != want {
			t.Errorf("%s: the error was %q, want %q", action, err.Error(), want)
		}
		if score, _ := e.World.Scoreboard.Score(uuid, "kills"); score != 5 {
			t.Errorf("%s: the score is %d, want the old one (5)", action, score)
		}
	}
}

// TestScoreboardPlayersDataTagParseError pins down the wording of a
// dataTag that cannot be parsed.
//
// (Measured on a real 1.12.2 server: `scoreboard players set
// @e[type=armor_stand] kills 4 {Tags:}` answers "Could not parse dataTag,
// reason: Expected value at: {Tags:<--[HERE]".)
func TestScoreboardPlayersDataTagParseError(t *testing.T) {
	e := newTestEngine(t)
	addArmourStand(t, e, "a1", "mytag")

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	err := runLine(e, "scoreboard players set @e[type=armor_stand] kills 4 {Tags:}")
	if err == nil {
		t.Fatal("expected an error for a dataTag that cannot be parsed")
	}
	want := "Could not parse dataTag, reason: Expected value at: {Tags:<--[HERE]"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}

// TestScoreboardPlayersDataTagNeedsAnEntity pins down that a holder that
// names no entity cannot carry a dataTag.
//
// 1.12.2 resolves the holder of the command as an entity when there is a
// dataTag to check, and a name that is no entity is reported as such.
//
// (Measured on a real 1.12.2 server: `scoreboard players set Steve kills 1
// {Tags:["mytag"]}` answers "Entity 'Steve' cannot be found", while the
// same command without the dataTag answers "Set score of kills for player
// Steve to 1".)
func TestScoreboardPlayersDataTagNeedsAnEntity(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	err := runLine(e, `scoreboard players set Steve kills 1 {Tags:["mytag"]}`)
	if err == nil {
		t.Fatal("expected an error for a holder that names no entity")
	}
	want := "Entity 'Steve' cannot be found"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	// Without the dataTag the name is a holder of its own.
	got, err := runLineFeedback(e, "scoreboard players set Steve kills 1")
	if err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	if len(got) != 1 || got[0] != "Set score of kills for player Steve to 1" {
		t.Errorf("the command said %v, want the score of Steve", got)
	}
}

// TestScoreboardPlayersDataTagIsCheckedForEveryHolder pins down that every
// entity of a target list is checked on its own.
//
// 1.12.2 expands a target list in the command dispatcher and runs the
// command once per matched entity, so every entity gets its own check and
// its own message -- and a failure of one of them does not stop the ones
// after it.
//
// (Measured on a real 1.12.2 server: with three armour stands, of which
// the first and the third carry the tag and the second does not, the
// command sets the score of the first one, answers "The dataTag does not
// match for <uuid of the second one>" and sets the score of the third one
// as well.)
func TestScoreboardPlayersDataTagIsCheckedForEveryHolder(t *testing.T) {
	e := newTestEngine(t)
	first := addArmourStand(t, e, "a1", "mytag")
	second := addArmourStand(t, e, "a2", "other")
	third := addArmourStand(t, e, "a3", "mytag")
	first.Pos = world.Vec{X: 0, Y: 0, Z: 0}
	second.Pos = world.Vec{X: 1, Y: 0, Z: 0}
	third.Pos = world.Vec{X: 2, Y: 0, Z: 0}

	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 3 {Tags:["mytag"]}`)
	if err == nil {
		t.Fatal("expected an error for the entity that does not match")
	}
	want := "The dataTag does not match for " + second.EnsureUUID()
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	// The entities that did match keep their score: the failure of the
	// one in between does not stop them.
	for _, ent := range []*world.Entity{first, third} {
		if score, tracked := e.World.Scoreboard.Score(ent.EnsureUUID(), "kills"); !tracked || score != 3 {
			t.Errorf("the score of %q is %d (tracked: %v), want 3",
				ent.EnsureUUID(), score, tracked)
		}
	}
	// The one that did not match has no score at all.
	if _, tracked := e.World.Scoreboard.Score(second.EnsureUUID(), "kills"); tracked {
		t.Errorf("the entity that did not match has a score")
	}
}

// TestScoreboardPlayersTagDataTag pins down the dataTag of a tag action.
//
// The tag action takes a dataTag as well, and it words a failure of it
// differently from the three actions that write a score.
//
// (Measured on a real 1.12.2 server: `scoreboard players tag
// @e[type=armor_stand] add t2 {Tags:["nope"]}` answers "The dataTag does
// not match for <uuid>" and `... add t3 {Tags:}` answers "Players tag
// command failed, reason: Expected value at: {Tags:<--[HERE]".)
func TestScoreboardPlayersTagDataTag(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1", "mytag")
	uuid := s.EnsureUUID()

	got, err := runLineFeedback(e,
		`scoreboard players tag @e[type=armor_stand] add t1 {Tags:["mytag"]}`)
	if err != nil {
		t.Fatalf("scoreboard players tag add: %v", err)
	}
	if len(got) != 1 || got[0] != "Tag t1 added" {
		t.Fatalf("the command said %v, want %q", got, "Tag t1 added")
	}

	err = runLine(e, `scoreboard players tag @e[type=armor_stand] add t2 {Tags:["nope"]}`)
	if err == nil {
		t.Fatal("expected an error for a dataTag that does not match")
	}
	want := "The dataTag does not match for " + uuid
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}

	err = runLine(e, `scoreboard players tag @e[type=armor_stand] add t3 {Tags:}`)
	if err == nil {
		t.Fatal("expected an error for a dataTag that cannot be parsed")
	}
	want = "Players tag command failed, reason: Expected value at: {Tags:<--[HERE]"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}

// TestScoreboardPlayersDataTagWording pins down the wording of the parse
// failures of a dataTag.
//
// Every wording was measured on a real 1.12.2 server.
func TestScoreboardPlayersDataTagWording(t *testing.T) {
	for _, tc := range []struct {
		dataTag string
		want    string
	}{
		{"{Tags:}", "Could not parse dataTag, reason: Expected value at: {Tags:<--[HERE]"},
		{"{Tags}", "Could not parse dataTag, reason: Expected ':' but got '}' at: {Tags}<--[HERE]"},
		{"}", "Could not parse dataTag, reason: Expected '{' but got '}' at: }<--[HERE]"},
		{"hello", "Could not parse dataTag, reason: Expected '{' but got 'h' at: h<--[HERE]"},
		{"5", "Could not parse dataTag, reason: Expected '{' but got '5' at: 5<--[HERE]"},
		{"{,}", "Could not parse dataTag, reason: Expected non-empty key at: {<--[HERE]"},
		{"{Tags:[1,2}", "Could not parse dataTag, reason: Expected ']' but got '}' at: {Tags:[1,2}<--[HERE]"},
		{"{Tags:1} extra", "Could not parse dataTag, reason: Trailing data found at: {Tags:1} e<--[HERE]"},
	} {
		e := newTestEngine(t)
		addArmourStand(t, e, "a1", "mytag")
		if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
			t.Fatalf("scoreboard objectives add: %v", err)
		}
		err := runLine(e, "scoreboard players set @e[type=armor_stand] kills 4 "+tc.dataTag)
		if err == nil {
			t.Errorf("%s: expected an error", tc.dataTag)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%s: the error was %q, want %q", tc.dataTag, err.Error(), tc.want)
		}
	}
}

// TestScoreboardPlayersDataTagThatTheLexerRejects pins down what happens
// to a dataTag whose braces do not balance.
//
// 1.12.2 splits the command into arguments first and parses the dataTag
// afterwards, so it reports "Could not parse dataTag, reason: ..." for
// such a dataTag. The runtime's lexer balances the braces while it reads
// the line and refuses the whole line when they do not match, so the
// command never runs.
//
// The wording of the failure is therefore the one of the lexer. It is
// pinned down here so that the difference is visible and not a surprise.
func TestScoreboardPlayersDataTagThatTheLexerRejects(t *testing.T) {
	for _, tc := range []struct {
		dataTag string
		want    string
	}{
		{"{Tags", "Syntax:1:1: unbalanced {"},
		{"{Tags:1", "Syntax:1:1: unbalanced {"},
		{"{", "Syntax:1:1: unbalanced {"},
		{"{Tags:1,", "Syntax:1:1: unbalanced {"},
		{`{Tags:"abc`, "Syntax:1:1: unterminated quoted string"},
	} {
		e := newTestEngine(t)
		addArmourStand(t, e, "a1", "mytag")
		if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
			t.Fatalf("scoreboard objectives add: %v", err)
		}
		err := runLine(e, "scoreboard players set @e[type=armor_stand] kills 4 "+tc.dataTag)
		if err == nil {
			t.Errorf("%s: expected an error", tc.dataTag)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%s: the error was %q, want %q", tc.dataTag, err.Error(), tc.want)
		}
	}
}

// TestMojangsonReason pins down the translation of the runtime's NBT
// parser failures into the reasons 1.12.2 reports.
//
// The wording of every reason was measured on a real 1.12.2 server.
func TestMojangsonReason(t *testing.T) {
	for _, tc := range []struct {
		text string
		want string
	}{
		{"{Tags:}", "Expected value at: {Tags:<--[HERE]"},
		{"{Tags}", "Expected ':' but got '}' at: {Tags}<--[HERE]"},
		{"{Tags", "Expected ':' but got '<EOF>' at: {Tags<--[HERE]"},
		{"{Tags:1", "Expected '}' but got '<EOF>' at: {Tags:1<--[HERE]"},
		{"{", "Expected '}' but got '<EOF>' at: {<--[HERE]"},
		{"{Tags:1,", "Expected key at: {Tags:1,<--[HERE]"},
		{"{,}", "Expected non-empty key at: {<--[HERE]"},
		{"{Tags:[1,2}", "Expected ']' but got '}' at: {Tags:[1,2}<--[HERE]"},
		{"{Tags:[1,2", "Expected ']' but got '<EOF>' at: {Tags:[1,2<--[HERE]"},
		{`{Tags:"abc`, "Missing termination quote at: {Tags:\"abc<--[HERE]"},
		{"{Tags:1} extra", "Trailing data found at: {Tags:1} e<--[HERE]"},
	} {
		_, err := parseDataTag(tc.text)
		if err == nil {
			t.Errorf("%s: expected an error", tc.text)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%s: the reason was %q, want %q", tc.text, err.Error(), tc.want)
		}
	}
}

// TestKillByUUID pins down that the UUID of an entity names it for the
// commands that resolve their target as an entity.
//
// 1.12.2 resolves such a target with `CommandAbstract.c`, which looks the
// name up as a player first and then parses it as a UUID.
//
// (Measured on a real 1.12.2 server: `kill <uuid>` kills the entity the
// UUID belongs to.)
func TestKillByUUID(t *testing.T) {
	e := newTestEngine(t)
	s := addArmourStand(t, e, "a1")
	uuid := s.EnsureUUID()

	if err := runLine(e, "kill "+uuid); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if e.World.EntityCount() != 0 {
		t.Errorf("the entity is still in the world: %d entities", e.World.EntityCount())
	}
}

// TestKillByAMalformedUUID pins down the wording of a target that is
// written like a UUID but is not one.
//
// (Measured on a real 1.12.2 server: `kill x-y-z-w-v` answers "The entity
// UUID 'x-y-z-w-v' is in an invalid format", while `kill a-b-c-d-e` --
// which parses as a UUID no entity has -- answers "Entity 'a-b-c-d-e'
// cannot be found".)
func TestKillByAMalformedUUID(t *testing.T) {
	e := newTestEngine(t)
	addArmourStand(t, e, "a1")

	err := runLine(e, "kill x-y-z-w-v")
	if err == nil {
		t.Fatal("expected an error for a malformed UUID")
	}
	want := "The entity UUID 'x-y-z-w-v' is in an invalid format"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}

	err = runLine(e, "kill a-b-c-d-e")
	if err == nil {
		t.Fatal("expected an error for a UUID no entity has")
	}
	want = "Entity 'a-b-c-d-e' cannot be found"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}

// TestScoreboardPlayersDataTagOfAnObjectiveThatIsMissing pins down that a
// missing objective is reported before a broken dataTag.
//
// 1.12.2 looks the objective up before it parses the dataTag.
//
// (Measured on a real 1.12.2 server: `scoreboard players set
// @e[type=armor_stand] nope 1 {Tags:}` answers "No objective was found by
// the name 'nope'".)
func TestScoreboardPlayersDataTagOfAnObjectiveThatIsMissing(t *testing.T) {
	e := newTestEngine(t)
	addArmourStand(t, e, "a1", "mytag")
	err := runLine(e, "scoreboard players set @e[type=armor_stand] nope 1 {Tags:}")
	if err == nil {
		t.Fatal("expected an error for an objective that does not exist")
	}
	want := "No objective was found by the name 'nope'"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}

// TestScoreboardPlayersTagNeedsAnEntity pins down that the holder of a tag
// action is always resolved as an entity.
//
// 1.12.2 does that before it looks at the dataTag, so a holder it cannot
// find is reported whether there is a dataTag or not.
//
// (Measured on a real 1.12.2 server: `scoreboard players tag Steve add x`
// answers "Entity 'Steve' cannot be found".)
func TestScoreboardPlayersTagNeedsAnEntity(t *testing.T) {
	for _, line := range []string{
		"scoreboard players tag Steve add marked",
		`scoreboard players tag Steve add marked {Tags:["mytag"]}`,
		"scoreboard players tag Steve remove marked",
		"scoreboard players tag Steve list",
	} {
		e := newTestEngine(t)
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q was expected to fail", line)
			continue
		}
		want := "Entity 'Steve' cannot be found"
		if err.Error() != want {
			t.Errorf("%q said %q, want %q", line, err.Error(), want)
		}
	}
}
