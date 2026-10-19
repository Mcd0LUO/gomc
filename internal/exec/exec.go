// Package exec implements the command execution engine.
//
// The engine is version-agnostic: it owns the execution context, the
// function call stack and the tick loop, while the actual command
// implementations are supplied by a version package through the Registry.
package exec

import (
	"fmt"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// Source describes who is running a command.
//
// It mirrors 1.12.2's ICommandListener closely enough to express the
// distinctions commands actually observe: the executor entity, the position
// the command runs at, the dimension, and whether feedback is sent.
type Source struct {
	// Name is the source's display name, e.g. "Server".
	Name string
	// Entity is the executing entity, or nil for the console.
	Entity *world.Entity
	// Pos is the position the command runs at.
	Pos world.Vec
	// Rot is the rotation the command runs with.
	Rot world.Rotation
	// World is the dimension the command runs in.
	World string
	// PermissionLevel is the source's permission level (0-4).
	PermissionLevel int
	// Feedback controls whether the source receives command output.
	Feedback bool
	// Silent marks a source whose output is suppressed.
	Silent bool
}

// Clone returns a copy of the source.
func (s *Source) Clone() *Source {
	cp := *s
	return &cp
}

// WithPos returns a copy of the source positioned at pos.
func (s *Source) WithPos(pos world.Vec) *Source {
	cp := s.Clone()
	cp.Pos = pos
	return cp
}

// WithEntity returns a copy of the source executed by entity.
func (s *Source) WithEntity(e *world.Entity) *Source {
	cp := s.Clone()
	cp.Entity = e
	return cp
}

// Context carries the state of a single command execution.
type Context struct {
	// Source is who is running the command.
	Source *Source
	// World is the world state being modified.
	World *world.World
	// Engine is the running engine.
	Engine *Engine
	// Command is the AST node being executed.
	Command *ast.Command
	// Depth is the current function call depth.
	Depth int
	// Args holds the command's parsed arguments.
	Args []ast.Arg
}

// Arg returns the argument at index i.
func (c *Context) Arg(i int) ast.Arg {
	if i < 0 || i >= len(c.Args) {
		return nil
	}
	return c.Args[i]
}

// StringArg returns the argument at index i as a string.
func (c *Context) StringArg(i int) string {
	if a, ok := c.Arg(i).(*ast.StringArg); ok {
		return a.Value
	}
	return ""
}

// Result is the outcome of running a command.
type Result struct {
	// Success reports whether the command succeeded.
	Success bool
	// Count is the command's success count, as vanilla reports it.
	Count int
}

// Error is an execution failure carrying an optional source position.
//
// Msg is what 1.12.2 would say. The remaining fields are gomc's own
// addition and are only rendered when the enhanced output is on.
type Error struct {
	// Msg explains the failure.
	Msg string
	// Line is the 1-based line of the failing command, or 0.
	Line int
	// Col is the 1-based column of the failing command, or 0.
	Col int
	// Path is the file the failing command came from, when known.
	Path string
	// Source is the text of the failing line, when known.
	Source string
	// Usage is the usage line of the command that failed, when known.
	Usage string
	// Hint is an extra suggestion.
	Hint string
}

func (e *Error) Error() string { return e.Msg }

// Errorf builds an execution error.
func Errorf(format string, args ...any) error {
	return &Error{Msg: fmt.Sprintf(format, args...)}
}

// Handler executes one command.
type Handler interface {
	// Name is the command's literal name.
	Name() string
	// Usage returns the usage string shown on a syntax error.
	Usage() string
	// Parse converts the arguments into a typed form and validates them.
	Parse(p *parser.Parser) ([]ast.Arg, error)
	// Run performs the command and returns its success count.
	Run(ctx *Context, args []ast.Arg) (int, error)
}

// Registry holds the commands available to the engine.
type Registry struct {
	handlers map[string]Handler
	order    []string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{handlers: map[string]Handler{}}
}

// Register adds a command handler.
func (r *Registry) Register(h Handler) {
	if _, ok := r.handlers[h.Name()]; !ok {
		r.order = append(r.order, h.Name())
	}
	r.handlers[h.Name()] = h
}

// Get returns the handler for a command name.
func (r *Registry) Get(name string) (Handler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

// Names returns the registered command names, sorted.
func (r *Registry) Names() []string {
	out := append([]string(nil), r.order...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// NBTOf is a helper to read an NBT argument from a parsed argument list.
func NBTOf(args []ast.Arg, i int) *nbt.Compound {
	if a, ok := args[i].(*ast.NBTArg); ok {
		return a.Value
	}
	return nil
}

// SelectorOf is a helper to read a selector argument from a parsed list.
func SelectorOf(args []ast.Arg, i int) *ast.Selector {
	if a, ok := args[i].(*ast.Selector); ok {
		return a
	}
	return nil
}
