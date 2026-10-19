package world

import "testing"

// 1.12.2 shows the translated name of a block, not its registry id, so
// "stone" comes out as "Stone" and "grass" as "Grass Block".
func TestBlockDisplayName(t *testing.T) {
	cases := map[string]string{
		"stone":            "Stone",
		"dirt":             "Dirt",
		"grass":            "Grass Block",
		"air":              "Air",
		"minecraft:stone":  "Stone",
		"polished_granite": "Polished Granite",
		"sandstone":        "Sandstone",
		"chest":            "Chest",
		// An unknown id still says something useful.
		"not_a_block": "Not A Block",
	}
	for name, want := range cases {
		if got := BlockDisplayName(name); got != want {
			t.Errorf("BlockDisplayName(%q) = %q, want %q", name, got, want)
		}
	}
}

// The names come from the language file, so the table has to actually
// contain the blocks a mcfunction is likely to touch.
func TestBlockDisplayNameCoversCommonBlocks(t *testing.T) {
	for _, name := range []string{"stone", "dirt", "sand", "air", "grass", "obsidian", "torch"} {
		if got := BlockDisplayName(name); got == name {
			t.Errorf("BlockDisplayName(%q) = %q, want a translated name", name, got)
		}
	}
}
