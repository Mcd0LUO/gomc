package v1_12_2

import (
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// locateCmd implements `/locate`.
//
// The 1.12.2 syntax is `/locate <feature>`: **one** argument, the name of a
// structure. The `<biome|structure> <name>` syntax belongs to 1.13; 1.12.2
// knows no biomes at all, and the names it offers are the eight structures
// its tab completion lists: Stronghold, Monument, Village, Mansion, EndCity,
// Fortress, Temple and Mineshaft.
//
// (Measured on a real 1.12.2 server: `locate biome plains` and
// `locate structure village` are both answered with
// "Usage: /locate <feature>", because the command reads exactly one
// argument.)
//
// The server answers with the position of the nearest structure of that
// name in the dimension the command runs in, and with a failure when there
// is none. gomc does not simulate world generation (see SPEC section 11),
// so its world holds no generated structures and there is never one to
// find: the command can only answer with the failure. The failure is the
// server's own wording, so the two agree on every input the server also
// fails on.
type locateCmd struct{}

// Name implements exec.Handler.
func (c *locateCmd) Name() string { return "locate" }

// Usage implements exec.Handler.
func (c *locateCmd) Usage() string { return locateUsageLine() }

// Parse implements exec.Handler.
//
// The server counts the arguments before it looks at any of them:
// `CommandLocate.execute` starts with `if (astring.length != 1)` and
// throws the usage of the command. The arguments are therefore collected
// first, and the count is checked when the command runs.
//
// The name is kept as it was written. The server looks it up in the map of
// the structure generators as it is, and that lookup is case sensitive.
//
// The name is read with the parser's string rule, which takes the quotes
// off a quoted name. 1.12.2 keeps them: it splits the command line on
// spaces and never looks at quotes, so `locate "Village"` is a search for
// the name `"Village"`. That difference belongs to the tokenizer, which
// every command shares, and not to this command.
func (c *locateCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	var args []ast.Arg
	for p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		args = append(args, &ast.StringArg{Value: name})
	}
	return args, nil
}

// Run implements exec.Handler.
//
// 1.12.2 refuses the command unless there is exactly one argument.
func (c *locateCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	if len(args) != 1 {
		return 0, locateUsageError()
	}
	name := args[0].(*ast.StringArg).Value

	// The server looks the name up in the map of the structure generators
	// of the dimension it runs in. A name that is not in the map, and a
	// structure that is out of reach, are both answered the same way:
	// "Unable to locate any <name> feature".
	//
	// gomc has no world generation, so its world has no structures at all
	// and the lookup can never succeed. The answer is the server's own
	// wording, which is also what the server says for a name it does not
	// know -- so gomc and the server agree on every input the server also
	// fails on.
	//
	// (Measured on a real 1.12.2 server: `locate plains` and
	// `locate bogus` are both answered with
	// "Unable to locate any <name> feature", the name being repeated as it
	// was written.)
	err := &exec.Error{Msg: fmt.Sprintf("Unable to locate any %s feature", name)}
	if locateOverworldStructures[name] {
		// The name is one the server would have looked for, so the two
		// differ here: the server would have answered with a position.
		// Say so, rather than let the difference pass unnoticed.
		err.Hint = "gomc does not simulate world generation, so it has no structures to locate"
	}
	return 0, err
}

// locateOverworldStructures are the structures 1.12.2 can find in the
// overworld.
//
// The server looks the name up in the map of the structure generators of
// the dimension the command runs in, and the map of the overworld holds
// exactly these six. EndCity belongs to the End and Fortress to the
// nether, so a `/locate` in the overworld never finds them.
//
// The lookup is case sensitive, so the names are written the way the server
// spells them.
//
// (Measured on a real 1.12.2 server: in the overworld `locate Stronghold`,
// `locate Village`, `locate Mineshaft`, `locate Temple`, `locate Monument`
// and `locate Mansion` all find a structure, while `locate EndCity` and
// `locate Fortress` are answered with
// "Unable to locate any <name> feature". The same names in lower case are
// not found either.)
var locateOverworldStructures = map[string]bool{
	"Stronghold": true,
	"Monument":   true,
	"Village":    true,
	"Mansion":    true,
	"Temple":     true,
	"Mineshaft":  true,
}

// locateUsageLine is the usage line 1.12.2 prints for `/locate`.
func locateUsageLine() string {
	if usage := commandUsage("locate"); usage != "" {
		return usage
	}
	return "/locate <feature>"
}

// locateUsageMessage is the message 1.12.2 answers a usage mistake with.
//
// The server sends `commands.generic.usage` with the usage line of the
// command, so the message is "Usage: <the line>".
func locateUsageMessage() string {
	prefix := commandUsage("generic")
	if prefix == "" {
		prefix = "Usage: %s"
	}
	return fmt.Sprintf(prefix, locateUsageLine())
}

// locateUsageError is the error 1.12.2 answers a usage mistake with.
func locateUsageError() error {
	return exec.Errorf("%s", locateUsageMessage())
}

// registerLocate adds the locate command.
//
// The permission level of the command is 2: 1.12.2 asks the command which
// level it needs (`CommandLocate.a`), and it answers 2.
func registerLocate(reg *exec.Registry) {
	reg.Register(&locateCmd{})
	reg.SetRequiredPermission("locate", 2)
}
