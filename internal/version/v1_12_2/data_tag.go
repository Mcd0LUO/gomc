package v1_12_2

import (
	"strings"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// dataTagText reads the dataTag of a command without parsing it.
//
// 1.12.2 splits the command into arguments first and parses the dataTag
// afterwards, so a dataTag that does not parse is the failure of the
// *command* and not of the reading of it: the target of the command is
// resolved first, and only then is the dataTag looked at.
//
// The dataTag is the **rest of the line**: the arguments from the
// dataTag's place on are joined with spaces and handed to the NBT parser
// (`CommandAbstract.a(listener, args, n)`), so a dataTag with spaces in it
// is one argument.
//
// (Measured on a real 1.12.2 server: `testfor @e[type=armor_stand {Tags:`
// answers "Entity '@e[type=armor_stand' cannot be found" and not a
// complaint about the dataTag.)
func dataTagText(p *parser.Parser) (string, error) {
	text := p.Rest()
	if strings.TrimSpace(text) == "" {
		return "", p.Errorf("expected NBT data but reached end of command")
	}
	return text, nil
}

// parseDataTagArg parses the dataTag of a command and reports a dataTag
// that does not parse the way 1.12.2 does.
//
// The wording is the one of the commands that take a dataTag without a
// scoreboard: "Data tag parsing failed: " followed by the reason the
// parser stopped at.
//
// (Measured on a real 1.12.2 server: `testfor @e[type=armor_stand] {Tags:`
// answers "Data tag parsing failed: Expected ':' but got '<EOF>' at:
// {Tags:<--[HERE]".)
func parseDataTagArg(text string) (*nbt.Compound, error) {
	compound, err := parseDataTag(text)
	if err != nil {
		return nil, exec.Errorf("Data tag parsing failed: %s", err)
	}
	return compound, nil
}
