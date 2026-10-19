// Package difftest implements the differential testing layer.
//
// The point of the layer is to answer one question: does gomc behave the
// same way a real Minecraft 1.12.2 server does? A case describes a
// scenario, running it produces a canonical observation, and the
// observation is compared against a golden file.
//
// Two scopes exist, because a real server can only report part of what
// gomc knows:
//
//   - ScopeServer covers what a server prints: the feedback of every
//     command and the error message of a failing one. This is the part
//     that can be captured from a real server's log, and it is therefore
//     the part that can be compared against one.
//   - ScopeFull adds the resulting world state: the scoreboard table, the
//     entities and the blocks. A 1.12.2 server cannot dump those from a
//     command, so this scope guards gomc against regressions rather than
//     against the server.
package difftest

import (
	"fmt"
	"strings"
)

// Case is one differential test case.
type Case struct {
	// Name identifies the case.
	Name string
	// Setup holds the commands that prepare the world.
	Setup []string
	// Run holds the commands under test.
	Run []string
}

// ParseCase reads a case from its textual form.
//
// The format is deliberately plain so that a case can be written by hand:
//
//	# a comment
//	name: scoreboard add
//
//	setup:
//	  scoreboard objectives add counter dummy
//
//	run:
//	  scoreboard players add Steve counter 2
func ParseCase(text string) (*Case, error) {
	c := &Case{}
	section := ""
	for i, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "name:"):
			c.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			continue
		case line == "setup:":
			section = "setup"
			continue
		case line == "run:":
			section = "run"
			continue
		}
		switch section {
		case "setup":
			c.Setup = append(c.Setup, line)
		case "run":
			c.Run = append(c.Run, line)
		default:
			return nil, fmt.Errorf("line %d: command %q is outside a setup or run section", i+1, line)
		}
	}
	if c.Name == "" {
		c.Name = "unnamed"
	}
	if len(c.Run) == 0 {
		return nil, fmt.Errorf("the case has no run section")
	}
	return c, nil
}

// Commands returns the setup and run commands in execution order.
func (c *Case) Commands() []string {
	out := make([]string, 0, len(c.Setup)+len(c.Run))
	out = append(out, c.Setup...)
	out = append(out, c.Run...)
	return out
}
