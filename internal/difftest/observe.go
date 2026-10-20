package difftest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/version/v1_12_2"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// Scope selects how much of an observation is rendered.
type Scope int

const (
	// ScopeServer renders only what a real server prints.
	ScopeServer Scope = iota
	// ScopeFull renders the server output and the resulting world state.
	ScopeFull
)

// StepResult is the outcome of a single command.
type StepResult struct {
	// Command is the command that was run.
	Command string
	// Count is the success count the command reported.
	Count int
	// Err is the error message, or empty when the command succeeded.
	Err string
	// Feedback holds the messages the command produced.
	Feedback []string
}

// Observation is the canonical record of a case run.
type Observation struct {
	// Case names the case.
	Case string
	// Steps holds one entry per command, setup first.
	Steps []StepResult
	// State is the rendered world state after the run.
	State string
}

// Run executes a case against a world and records what happened.
//
// The world is loaded first so the case can call the functions it needs,
// exactly as a server would have them available.
func Run(root string, c *Case) (*Observation, error) {
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	engine := exec.NewEngine(reg, world.NewWorld())
	engine.SetRoot(root)
	// A broken function elsewhere in the world must not stop the case.
	engine.LoadFunctions()

	src := &exec.Source{
		Name:            "Server",
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}

	obs := &Observation{Case: c.Name}
	for _, cmd := range c.Commands() {
		before := len(engine.Feedback)
		count, err := engine.ExecuteLine(cmd, 1, src)
		// The `sendCommandFeedback` game rule decides whether the
		// messages are shown, exactly as on a real server.
		var feedback []string
		if engine.World.Rules.Bool("sendCommandFeedback") {
			feedback = append(feedback, engine.Feedback[before:]...)
		}
		engine.Feedback = engine.Feedback[:before]

		res := StepResult{Command: cmd, Count: count, Feedback: feedback}
		if err != nil {
			res.Err = errorMessage(err)
		}
		// A command that fails inside a function does not stop the
		// function, so those failures are collected by the engine.
		for _, e := range engine.TakeErrors() {
			if res.Err == "" {
				res.Err = e.Error()
			} else {
				res.Err += "; " + e.Error()
			}
		}
		obs.Steps = append(obs.Steps, res)
	}
	obs.State = RenderState(engine.World)
	return obs, nil
}

// errorMessage returns the message of an error without the position.
//
// The scope of this layer is what a **server** prints, and a server prints
// the message alone. The position is gomc's own addition: the command line
// shows it next to the message, so repeating it inside the message would
// only make the two harder to compare.
func errorMessage(err error) string {
	if pe, ok := err.(*parser.Error); ok {
		return pe.Msg
	}
	return err.Error()
}

// Render returns the canonical text form of an observation.
//
// The output is stable and diffable, so a golden file stays meaningful
// across runs.
func (o *Observation) Render(scope Scope) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# case: %s\n", o.Case)
	for i, step := range o.Steps {
		fmt.Fprintf(&sb, "step %d: %s\n", i+1, step.Command)
		if step.Err != "" {
			fmt.Fprintf(&sb, "  error: %s\n", step.Err)
		} else {
			fmt.Fprintf(&sb, "  count: %d\n", step.Count)
		}
		for _, msg := range step.Feedback {
			fmt.Fprintf(&sb, "  feedback: %s\n", msg)
		}
	}
	if scope == ScopeFull {
		sb.WriteString(o.State)
	}
	return sb.String()
}

// RenderState returns the canonical text form of a world's state.
//
// This is the part a 1.12.2 server cannot produce on demand, so it guards
// gomc against regressions instead of against the server.
func RenderState(w *world.World) string {
	var sb strings.Builder
	sb.WriteString("state:\n")
	sb.WriteString("  scores:\n")
	for _, line := range renderScores(w) {
		fmt.Fprintf(&sb, "    %s\n", line)
	}
	sb.WriteString("  entities:\n")
	for _, line := range renderEntities(w) {
		fmt.Fprintf(&sb, "    %s\n", line)
	}
	sb.WriteString("  blocks:\n")
	for _, line := range renderBlocks(w) {
		fmt.Fprintf(&sb, "    %s\n", line)
	}
	return sb.String()
}

// renderScores renders the scoreboard table.
func renderScores(w *world.World) []string {
	var out []string
	for _, obj := range w.Scoreboard.Objectives() {
		for _, holder := range w.Scoreboard.Holders(obj.Name) {
			score, ok := w.Scoreboard.Score(holder, obj.Name)
			if !ok {
				continue
			}
			out = append(out, fmt.Sprintf("%s/%s = %d", obj.Name, holder, score))
		}
	}
	sort.Strings(out)
	return out
}

// renderEntities renders the entity table.
func renderEntities(w *world.World) []string {
	var out []string
	for _, e := range w.EntityList() {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%s %s @ %.2f,%.2f,%.2f", e.ID, e.Type, e.Pos.X, e.Pos.Y, e.Pos.Z)
		if e.NBT != nil && e.NBT.Len() > 0 {
			fmt.Fprintf(&sb, " nbt=%s", e.NBT.String())
		}
		if tags := e.TagList(); len(tags) > 0 {
			fmt.Fprintf(&sb, " tags=%s", strings.Join(tags, ","))
		}
		out = append(out, sb.String())
	}
	return out
}

// renderBlocks renders the modified blocks.
func renderBlocks(w *world.World) []string {
	positions := w.Blocks.Positions()
	sort.Slice(positions, func(i, j int) bool {
		a, b := positions[i], positions[j]
		if a.X != b.X {
			return a.X < b.X
		}
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		return a.Z < b.Z
	})
	var out []string
	for _, p := range positions {
		st := w.Blocks.Get(p)
		out = append(out, fmt.Sprintf("%d,%d,%d = %s %d", p.X, p.Y, p.Z, st.Name, st.Data))
	}
	return out
}
