package v1_12_2

import (
	"fmt"
	"strconv"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// The bounds 1.12.2 puts on a coordinate.
//
// `CommandAbstract.b(base, text, min, max, flag)` reads a coordinate with
// those bounds, and which bounds a coordinate gets depends on what the
// coordinate is for:
//
//   - a block position: X and Z are limited to [-30000000, 30000000] and Y
//     to [0, 256] (`CommandAbstract.a(listener, args, start, centerCorrect)`);
//   - a plain number: all three are limited to [-30000000, 30000000]
//     (`CommandAbstract.b(base, text, flag)`, used by `/execute`,
//     `/particle`, `/playsound`, `/summon`…);
//   - the position of `/tp`: X and Z are limited to [-30000000, 30000000]
//     and Y to [-4096, 4096].
//
// A coordinate outside the bounds is refused with
// `commands.generic.num.tooBig` / `commands.generic.num.tooSmall`, and the
// number in the message is formatted with "%.2f".
//
// (Measured on a real 1.12.2 server: `setblock 30000001 4 0 stone` answers
// "The number you have entered (30000001.00) is too big, it must be at most
// 30000000" and `setblock 0 -1 0 stone` answers "The number you have
// entered (-1.00) is too small, it must be at least 0".)
const (
	coordMinXZ = -30000000
	coordMaxXZ = 30000000
	coordMinY  = 0
	coordMaxY  = 256
	// tpMinY and tpMaxY are the bounds of the Y of `/tp`.
	tpMinY = -4096
	tpMaxY = 4096
)

// The coordinate specs 1.12.2 uses.
//
// Every command that reads a coordinate picks one of these. A spec says
// which bounds each of the three axes gets, and whether a whole number
// names the middle of a block.
var (
	// coordBlockPos is the spec of a block position: X and Z are limited
	// to ±30000000 and Y to [0, 256].
	coordBlockPos = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
		Y: parser.CoordAxis{Min: coordMinY, Max: coordMaxY},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
	}
	// coordBlockPosCentered is the spec of a block position that is read
	// in the "centre" mode, like `/spawnpoint` reads it: a whole number
	// names the middle of a block.
	//
	// Only the X and Z are centred: 1.12.2 reads the Y of a block
	// position with `b(base.getY(), text, 0, 256, false)`, and the
	// `false` is the "centre" flag.
	coordBlockPosCentered = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
		Y: parser.CoordAxis{Min: coordMinY, Max: coordMaxY},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
	}
	// coordLoose is the spec of a plain number: all three axes are
	// limited to ±30000000. It is what `/execute` reads its
	// coordinates with.
	coordLoose = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
		Y: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
	}
	// coordLooseCentered is the spec of a plain number in the "centre"
	// mode, like `/particle` and `/playsound` read theirs.
	coordLooseCentered = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
		Y: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
	}
	// coordSummon is the spec of `/summon`: X and Z are read in the
	// "centre" mode, Y is not, and all three are limited to ±30000000.
	coordSummon = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
		Y: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
	}
	// coordTp is the spec of `/tp`: X and Z are read in the "centre"
	// mode and limited to ±30000000, Y is limited to ±4096.
	coordTp = parser.CoordSpec{
		X: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
		Y: parser.CoordAxis{Min: tpMinY, Max: tpMaxY},
		Z: parser.CoordAxis{Min: coordMinXZ, Max: coordMaxXZ, Center: true},
	}
)

// checkResolvedCoord checks a coordinate that has already been resolved
// against the bounds of spec.
//
// 1.12.2 checks a coordinate where it reads it, and `/tp` reads its
// coordinates **inside** `CommandTeleport.execute` with
// `CommandAbstract.b(listener, text, min, max, flag)`. The base of a
// relative coordinate is therefore the listener's position at the moment
// the command runs, and the number that is checked is `base + offset` --
// the final coordinate, not the literal. Nothing further down the path
// checks it: `Entity.setPosition` clamps nothing, so a coordinate that
// slips past the read is written to the entity as it is.
//
// gomc parses a function's lines once, when the function is loaded, and
// the position the function will run at is not known then
// (`parser.CoordBase.Known` is false), so the parse-time check only sees
// the literal. The check here is the one that matches the server for such
// a line: it runs after the base has been added, and it is what makes a
// relative `/tp` inside a function that lands outside ±30000000 fail
// instead of moving the entity there.
//
// The bounds are checked X, then Y, then Z, which is the order
// `CommandTeleport.execute` reads them in, so the axis a message names is
// the first one that is out of bounds. A coordinate outside the bounds is
// refused with `commands.generic.num.tooBig` / `tooSmall`, and the number
// in the message is formatted with "%.2f" -- the same text the parser
// produces.
func checkResolvedCoord(spec parser.CoordSpec, x, y, z float64) error {
	for _, a := range [...]struct {
		value float64
		axis  parser.CoordAxis
	}{{x, spec.X}, {y, spec.Y}, {z, spec.Z}} {
		if a.value < a.axis.Min {
			return exec.Errorf("The number you have entered (%s) is too small, it must be at least %s",
				fmt.Sprintf("%.2f", a.value), coordBoundText(a.axis.Min))
		}
		if a.value > a.axis.Max {
			return exec.Errorf("The number you have entered (%s) is too big, it must be at most %s",
				fmt.Sprintf("%.2f", a.value), coordBoundText(a.axis.Max))
		}
	}
	return nil
}

// coordBoundText renders a bound the way the server puts it into its
// message: the bound is an int there.
func coordBoundText(v float64) string {
	return strconv.FormatInt(int64(v), 10)
}
