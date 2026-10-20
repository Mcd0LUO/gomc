package v1_12_2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/world"
)

// The progress file of the advancements of a player.
//
// 1.12.2 keeps the progress of a player in
// `<world>/advancements/<uuid>.json` -- `PlayerList` builds the path out
// of the directory of the world, the string "advancements" and the uuid
// of the player. The file is a JSON object that maps the name of an
// advancement to its progress:
//
//	{
//	  "minecraft:story/root": {
//	    "criteria": {
//	      "crafting_table": "2026-10-20 21:44:38 +0800"
//	    },
//	    "done": true
//	  }
//	}
//
// The value of a criterion is the moment the criterion was completed. The
// server writes it with `SimpleDateFormat("yyyy-MM-dd HH:mm:ss Z")` and
// reads it back with the same format. The "done" flag is not read back:
// the server works it out from the requirements of the advancement.
//
// The directory is the one the custom advancements live in, but the two
// do not collide. A custom advancement is in a directory named after its
// namespace (`<world>/advancements/<namespace>/<path>.json`), and the
// server skips every file that is not in such a directory; the progress
// of a player is a file right in the directory
// (`<world>/advancements/<uuid>.json`).
//
// (Measured on a real 1.12.2 server: the file of a player holds the
// advancements the player has started, with the moment every criterion
// was completed and the "done" flag of the advancement.)

// advancementSaveTime is the moment gomc writes into a progress file.
//
// 1.12.2 writes the moment the criterion was completed, but gomc is a
// deterministic command interpreter: the same commands have to produce the
// same files every time. The value is therefore a constant. The value
// itself does not matter -- the server only shows it, and gomc only looks
// at whether a criterion is there at all.
//
// (The format is the one of the server: `SimpleDateFormat("yyyy-MM-dd
// HH:mm:ss Z")`.)
const advancementSaveTime = "1970-01-01 00:00:00 +0000"

// advancementSaveEntry is the progress of one advancement.
//
// It is the shape of the value of the progress file of a player.
type advancementSaveEntry struct {
	// Criteria holds the criteria the player has completed, by name. The
	// value of a criterion is the moment it was completed.
	Criteria map[string]string `json:"criteria,omitempty"`
	// Done tells whether the advancement is complete.
	Done bool `json:"done"`
}

// advancementProgressPath returns the file the progress of a player is
// kept in.
//
// The name of the file is the uuid of the player. The real server draws
// that uuid at random, so gomc cannot know it -- it *derives* the uuid
// from the name of the player instead (`world.DeriveEntityUUID`). The
// derived uuid is stable, and it is not the uuid the real server uses, so
// the file of gomc never overwrites the file of the real server (and the
// other way round).
func advancementProgressPath(root, player string) string {
	return filepath.Join(root, customAdvancementDir, world.DeriveEntityUUID(player)+".json")
}

// readAdvancementProgress reads the progress files of a world.
//
// The files are the ones right in `<world>/advancements`: the server
// names them after the uuid of the player. The progress comes back keyed
// by that uuid, so the player it belongs to can be found later.
//
// A file that does not parse is reported and skipped, the way the server
// reports it and keeps the progress of the other players.
func readAdvancementProgress(root string) (map[string]map[string]map[string]bool, []error) {
	if root == "" {
		return nil, nil
	}
	base := filepath.Join(root, customAdvancementDir)
	entries, err := os.ReadDir(base)
	if err != nil {
		// A world without a progress file has no progress, and the
		// server does not say anything about it.
		return nil, nil
	}
	out := map[string]map[string]map[string]bool{}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		uuid := strings.TrimSuffix(name, ".json")
		if !isAdvancementUUID(uuid) {
			// A file that is not named after a uuid is not the
			// progress of a player.
			continue
		}
		progress, err := readAdvancementProgressFile(filepath.Join(base, name))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		out[uuid] = progress
	}
	return out, errs
}

// readAdvancementProgressFile reads one progress file.
//
// The "done" flag of the file is not read: the server works it out from
// the requirements of the advancement, and so does gomc.
func readAdvancementProgressFile(path string) (map[string]map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read the advancement progress from %s: %w", path, err)
	}
	var file map[string]advancementSaveEntry
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing error loading the advancement progress from %s: %w", path, err)
	}
	out := map[string]map[string]bool{}
	for name, entry := range file {
		if len(entry.Criteria) == 0 {
			continue
		}
		criteria := map[string]bool{}
		for criterion := range entry.Criteria {
			criteria[criterion] = true
		}
		out[name] = criteria
	}
	return out, nil
}

// writeAdvancementProgress writes the progress of a player.
//
// The file is the one the server would write: a JSON object that maps the
// name of an advancement to the criteria the player has completed and
// whether the advancement is complete. Only the advancements the player
// has started are in it, the way the server only saves the advancements
// that have a criterion.
//
// The names are sorted, so the same progress always produces the same
// file. (The server writes them in the order of its hash map, which is
// stable for the same names but not sorted.)
func writeAdvancementProgress(path string, set *advancementSet, progress map[string]map[string]bool) error {
	file := map[string]advancementSaveEntry{}
	for name, criteria := range progress {
		if len(criteria) == 0 {
			continue
		}
		advancement, ok := set.lookup(name)
		if !ok {
			// The server only knows the advancements of its
			// registry, so an advancement that is not in it any
			// more is not saved.
			continue
		}
		entry := advancementSaveEntry{Criteria: map[string]string{}}
		for criterion := range criteria {
			entry.Criteria[criterion] = advancementSaveTime
		}
		entry.Done = advancement.done(criteria)
		file[name] = entry
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// isAdvancementUUID reports whether the text has the shape of a uuid.
//
// The progress file of a player is named after the uuid of the player, so
// this tells the progress files apart from the other files of the
// directory.
func isAdvancementUUID(text string) bool {
	if len(text) != 36 {
		return false
	}
	for i, c := range text {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHexDigit(c) {
				return false
			}
		}
	}
	return true
}

// isHexDigit reports whether the character is a hexadecimal digit.
func isHexDigit(c rune) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		return true
	}
	return false
}
