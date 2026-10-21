package v1_12_2

import (
	"encoding/json"
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// jsonArg reads the JSON message of a command and checks that it is JSON.
//
// 1.12.2 hands the text to Gson (`ChatSerializer.a`) and reports a broken
// one with "Invalid json: " followed by the reason the parser stopped at.
// The reason is Gson's, and gomc does not repeat it word for word.
//
// The failure does **not** stop the command while it is parsed: 1.12.2
// resolves the target of the command before it looks at the JSON, so a
// command whose target does not exist fails with the target and not with
// the JSON. The error therefore travels with the argument, and it is
// reported when the command runs.
//
// (Measured on a real 1.12.2 server: `title @s title "e f"` answers "Player
// '@s' cannot be found" and not a complaint about the JSON.)
func jsonArg(p *parser.Parser) (*ast.JSONArg, error) {
	line, col, offset := p.Pos()
	arg, err := p.JSON()
	if err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal([]byte(arg.Value), &value); err != nil {
		arg.Err = &parser.Error{
			Kind:   parser.KindSyntax,
			Msg:    fmt.Sprintf("Invalid json: %s", err),
			Line:   line,
			Col:    col,
			Offset: offset,
		}
	}
	return arg, nil
}

// jsonArgError returns the failure of a JSON argument, if it has one.
//
// It is the error the command reports when it runs.
func jsonArgError(args []ast.Arg, index int) error {
	if index >= len(args) {
		return nil
	}
	j, ok := args[index].(*ast.JSONArg)
	if !ok {
		return nil
	}
	return j.Err
}
