package v1_12_2

import (
	"testing"
)

// notSummonableTypes are the names `/summon` refuses on the real server.
//
// The four fireballs and the thrown potion are registered entity types,
// but the server only builds them as the shot of something else, and the
// player is not in the entity registry at all. The server reports all of
// them the same way.
var notSummonableTypes = []string{
	"player",
	"potion",
	"fireball",
	"small_fireball",
	"wither_skull",
	"dragon_fireball",
}

// TestSummonEveryKnownType summons every type the runtime knows.
//
// 1.12.2 knows 85 entity type names and can bring 79 of them into the
// world; the other six are refused (see notSummonableTypes).
func TestSummonEveryKnownType(t *testing.T) {
	e := newTestEngine(t)
	names := e.World.EntityTypes.Names()
	if len(names) != 80 {
		t.Fatalf("the runtime knows %d entity types, want 80", len(names))
	}
	for _, name := range names {
		if name == "player" {
			continue
		}
		line := "summon " + name + " 0 4 0"
		if name == "item" {
			// The item entity is the one type that needs more than
			// its name; see TestSummonItemNeedsItsStack.
			line = `summon item 0 4 0 {Item:{id:"stone",Count:1b}}`
		}
		if err := runLine(e, line); err != nil {
			t.Errorf("summon %s: %v", name, err)
		}
	}
}

// TestSummonItemNeedsItsStack pins down the rule for the item entity.
//
// 1.12.2's entity factory only accepts an item entity that carries a
// non-empty item stack. An empty stack makes the entity die right after
// it is created, so the server refuses to bring it into the world.
//
// (Measured on a real 1.12.2 server: `summon item` and `summon item
// {Item:{}}` answer "Unable to summon object", and so does `summon item
// {Item:{id:"stone"}}` — the count is what decides, because 1.12.2
// reads it with `getByte`. `summon item {Item:{id:"stone",Count:1b}}`
// works.)
func TestSummonItemNeedsItsStack(t *testing.T) {
	e := newTestEngine(t)
	for _, line := range []string{
		"summon item 0 4 0",
		"summon item 0 4 0 {Item:{}}",
		"summon item 0 4 0 {Item:{id:\"stone\"}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:0b,Slot:0b}}",
		// getByte truncates to a signed byte, so 256 becomes 0 and -1
		// stays negative.
		"summon item 0 4 0 {Item:{id:\"stone\",Count:256}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:-1}}",
	} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: expected a failure", line)
			continue
		}
		if got := err.Error(); got != "Unable to summon object" {
			t.Errorf("%q: message = %q, want %q", line, got, "Unable to summon object")
		}
	}
	// With a stack it works, and the count may be any number type: 1.12.2
	// reads it with getByte, so a plain int counts as one.
	for _, line := range []string{
		"summon item 0 4 0 {Item:{id:\"stone\",Count:1b}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:1}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:1s}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:1L}}",
		"summon item 0 4 0 {Item:{id:\"stone\",Count:1.0f}}",
	} {
		feedback, err := runLineFeedback(e, line)
		if err != nil {
			t.Errorf("%q: %v", line, err)
			continue
		}
		if len(feedback) != 1 || feedback[0] != "Object successfully summoned" {
			t.Errorf("%q: feedback = %v, want [Object successfully summoned]", line, feedback)
		}
	}
}

// TestSummonTypeCount pins the number of types `/summon` accepts.
//
// The count was measured on the real server: of the 85 names it knows,
// 78 are summoned by the bare command, the item entity needs its item
// stack, and the six names of notSummonableTypes are refused.
func TestSummonTypeCount(t *testing.T) {
	e := newTestEngine(t)
	count := 0
	for _, name := range e.World.EntityTypes.Names() {
		if name == "player" {
			continue
		}
		count++
	}
	if count != 79 {
		t.Errorf("/summon accepts %d entity types, want 79", count)
	}
}

// TestSummonRefusesTheTypesTheServerRefuses checks the failure text.
//
// 1.12.2 does not check the type name at all: it puts the raw text into
// the NBT "id" and lets the entity factory build the entity. Every name
// the factory cannot build is reported the same way.
func TestSummonRefusesTheTypesTheServerRefuses(t *testing.T) {
	e := newTestEngine(t)
	names := append([]string{}, notSummonableTypes...)
	// A name the registry does not have at all.
	names = append(names, "nosuchentity")
	// The old spelling of the renamed zombie pigman.
	names = append(names, "pigzombie")
	// A name of another namespace: the registry holds "minecraft" only.
	names = append(names, "foo:zombie")
	for _, name := range names {
		err := runLine(e, "summon "+name+" 0 4 0")
		if err == nil {
			t.Errorf("summon %s: expected a failure", name)
			continue
		}
		if got := err.Error(); got != "Unable to summon object" {
			t.Errorf("summon %s: message = %q, want %q", name, got, "Unable to summon object")
		}
	}
}

// TestSummonAcceptsTheMinecraftNamespace checks the spellings of the type
// name that name the same type on the real server.
//
// 1.12.2 builds a `MinecraftKey` from the text: the namespace is only
// read from the name when the colon is neither the first nor the second
// character, and the name is lower-cased.
func TestSummonAcceptsTheMinecraftNamespace(t *testing.T) {
	e := newTestEngine(t)
	for _, name := range []string{
		"zombie",
		"minecraft:zombie",
		":zombie",
		"a:zombie",
		"ZOMBIE",
		"Zombie",
	} {
		if err := runLine(e, "summon "+name+" 0 4 0"); err != nil {
			t.Errorf("summon %s: %v", name, err)
		}
	}
}

// TestSummonCreatesTheEntity checks that a summoned entity is in the
// world and that the command reports the success.
func TestSummonCreatesTheEntity(t *testing.T) {
	e := newTestEngine(t)
	feedback, err := runLineFeedback(e, "summon zombie 0 4 0")
	if err != nil {
		t.Fatalf("summon: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "Object successfully summoned" {
		t.Fatalf("feedback = %v, want [Object successfully summoned]", feedback)
	}
	found := false
	for _, ent := range e.World.EntityList() {
		if ent.Type != "zombie" {
			continue
		}
		found = true
		if ent.Pos.X != 0 || ent.Pos.Y != 4 || ent.Pos.Z != 0 {
			t.Errorf("the entity is at %v, want (0, 4, 0)", ent.Pos)
		}
	}
	if !found {
		t.Error("the zombie is not in the world")
	}
}
