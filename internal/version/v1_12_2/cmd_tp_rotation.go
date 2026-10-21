package v1_12_2

import (
	"math"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// The rotation of an entity.
//
// 1.12.2 reads the yaw and the pitch of a `/tp` the way it reads a
// coordinate: a number is an absolute angle, a number behind a tilde is
// measured from the angle the entity has, and a bare tilde keeps the
// angle the entity has. An argument that is not there at all keeps the
// angle as well -- so `tp 1 2 3 90` turns the entity to the yaw 90 and
// leaves its pitch alone.
//
// The two angles are then put into their range: the yaw is folded into
// [-180, 180) and the pitch is clamped to [-90, 90].
//
// (`CommandTp` reads them with
// `CommandTp.a(entity.yaw, text, false)` and
// `CommandTp.a(entity.pitch, text, false)`, and then puts them through
// `MathHelper.g` and `MathHelper.a(value, -90.0f, 90.0f)`.)
type tpRotation struct {
	// yaw is the angle the entity turns to, in degrees.
	yaw float64
	// pitch is the angle the entity looks up or down, in degrees.
	pitch float64
	// yawGiven marks that the command named a yaw.
	yawGiven bool
	// yawRelative marks that the yaw is measured from the angle the
	// entity has.
	yawRelative bool
	// pitchGiven marks that the command named a pitch.
	pitchGiven bool
	// pitchRelative marks that the pitch is measured from the angle the
	// entity has.
	pitchRelative bool
}

// parseTPRotation reads the yaw and the pitch of a `/tp`.
//
// 1.12.2 gives both of them the same default: a bare tilde, which keeps
// the angle the entity has. The yaw is read first and the pitch second,
// and either of them may be left out.
func parseTPRotation(p *parser.Parser) (tpRotation, error) {
	var rot tpRotation
	if p.More() {
		yaw, rel, err := parseAngle(p)
		if err != nil {
			return rot, err
		}
		rot.yaw, rot.yawRelative, rot.yawGiven = yaw, rel, true
	}
	if p.More() {
		pitch, rel, err := parseAngle(p)
		if err != nil {
			return rot, err
		}
		rot.pitch, rot.pitchRelative, rot.pitchGiven = pitch, rel, true
	}
	return rot, nil
}

// parseAngle reads one angle of a `/tp`.
//
// The text is a number, or a number behind a tilde. A bare tilde is zero
// and relative, which is how 1.12.2 spells "keep what you have".
func parseAngle(p *parser.Parser) (float64, bool, error) {
	t, ok := p.Next()
	if !ok {
		return 0, false, p.Errorf("expected a number but reached end of command")
	}
	raw := t.Text
	rel := strings.HasPrefix(raw, "~")
	if rel {
		raw = raw[1:]
	}
	if raw == "" {
		return 0, true, nil
	}
	f, ok := parser.JavaDouble(raw)
	if !ok {
		return 0, false, p.ErrorAt(t, "'%s' is not a valid number", t.Text)
	}
	return f, rel, nil
}

// applyTPRotation puts the rotation on an entity.
//
// The yaw is measured from the south and grows clockwise, and it is
// folded into [-180, 180). The pitch is clamped to [-90, 90], because
// that is the range an entity can look in.
//
// An angle the command did not name keeps the angle the entity has, and
// a tilde is measured from it.
func applyTPRotation(e *world.Entity, rot tpRotation) {
	if rot.yawGiven {
		yaw := rot.yaw
		if rot.yawRelative {
			yaw += e.Rot.Yaw
		}
		e.Rot.Yaw = foldAngle(yaw)
	}
	if rot.pitchGiven {
		pitch := rot.pitch
		if rot.pitchRelative {
			pitch += e.Rot.Pitch
		}
		// The pitch is folded into [-180, 180) **first** and clamped
		// to [-90, 90] after that, so a pitch of 200 ends up at -90 and
		// not at 90.
		//
		// (Measured on a real 1.12.2 server, probe case
		// 60-tp-rotation-bounds: `tp @e[type=armor_stand] 1 4 1 0 200`
		// leaves the entity at a pitch of -90 -- the selector
		// `rx=-90` finds it and `rx=90` does not. `CommandTp` does
		// the same: `MathHelper.g` and then
		// `MathHelper.a(value, -90.0f, 90.0f)`.)
		e.Rot.Pitch = clampAngle(foldAngle(pitch))
	}
}

// foldAngle puts an angle into [-180, 180).
//
// It is `MathHelper.g`, which the server uses for the yaw.
func foldAngle(angle float64) float64 {
	a := math.Mod(angle, 360)
	if a >= 180 {
		a -= 360
	}
	if a < -180 {
		a += 360
	}
	return a
}

// clampAngle puts an angle into [-90, 90].
//
// It is `MathHelper.a(value, -90.0f, 90.0f)`, which the server uses for
// the pitch.
func clampAngle(angle float64) float64 {
	if angle > 90 {
		return 90
	}
	if angle < -90 {
		return -90
	}
	return angle
}
