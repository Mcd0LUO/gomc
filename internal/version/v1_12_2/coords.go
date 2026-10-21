package v1_12_2

import "github.com/Mcd0LUO/gomc/internal/parser"

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
