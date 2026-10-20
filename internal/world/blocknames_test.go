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

// 1.12.2 reuses some of the names of 1.11 for a different block, and it
// still takes the old names. The table of the registry names comes first,
// so the meaning of 1.12.2 wins.
//
// (Measured on a real 1.12.2 server: "stonebrick" is the stone bricks
// there, while in 1.11 it was the cobblestone.)
func TestBlockDisplayNamePrefersTheRegistryName(t *testing.T) {
	cases := map[string]string{
		// 1.12.2 reuses the name, and it means the stone bricks now.
		"stonebrick": "Stone Bricks",
		// The names of 1.12.
		"end_stone":    "End Stone",
		"cobblestone":  "Cobblestone",
		"hay_block":    "Hay Bale",
		"carpet":       "Carpet",
		"nether_brick": "Nether Brick",
		// The names of 1.11 still work, and they are looked up in the
		// table of the older names.
		"whiteStone": "End Stone",
		"doorWood":   "Wooden Door",
		"woolCarpet": "Carpet",
		"hayBlock":   "Hay Bale",
	}
	for name, want := range cases {
		if got := BlockDisplayName(name); got != want {
			t.Errorf("BlockDisplayName(%q) = %q, want %q", name, got, want)
		}
	}
}
