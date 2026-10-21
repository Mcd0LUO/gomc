// Package exec implements the command execution engine.
//
// The engine is version-agnostic: it owns the execution context, the
// function call stack and the tick loop, while the actual command
// implementations are supplied by a version package through the Registry.
package exec

import (
	"fmt"
	"sort"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// Source describes who is running a command.
//
// It mirrors 1.12.2's ICommandListener closely enough to express the
// distinctions commands actually observe: the executor entity, the position
// the command runs at, the dimension, and whether feedback is sent.
type Source struct {
	// Name is the source's display name, e.g. "Server".
	Name string
	// Entity is the entity the command runs as, or nil for the console.
	//
	// It is the **executor** of the command: `@s` refers to it and the
	// output of the command is prefixed with its name.
	Entity *world.Entity
	// Pos is the position the command runs at.
	Pos world.Vec
	// ListenerPlayer reports whether the **listener** of the command is a
	// player.
	//
	// The command runs *as* an entity and *for* a listener. In 1.12.2 the
	// two are usually the same, but the command of an `/execute` runs as
	// the target entity while the listener is the wrapper of the
	// execute. The listener is what decides two things:
	//
	//   - the base a relative coordinate is measured from: a player is
	//     the listener of their own commands, and their chunk
	//     coordinates are half a block higher than the block they
	//     stand in;
	//   - whether the commands that need a player find one.
	//
	// It is false for the console and for the command of an `/execute`.
	ListenerPlayer bool
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

// RelativeBase returns the position a relative coordinate is measured
// from.
//
// 1.12.2 reads a relative coordinate against
// `ICommandListener.getChunkCoordinates()`, and that is always a **block
// position**:
//
//   - the console is at (0, 0, 0);
//   - a player is at (⌊x⌋, ⌊y+0.5⌋, ⌊z⌋) — the Y is half a block up,
//     because that is the block the body of the player is in
//     (`EntityPlayer.getChunkCoordinates`);
//   - the command of an `/execute` is at the block the execute position
//     points at (`CommandListenerWrapper`).
func (s *Source) RelativeBase() world.Vec {
	if s.ListenerPlayer && s.Entity != nil {
		// The player is the listener: their chunk coordinates are half a
		// block higher than the block they stand in.
		p := world.Vec{X: s.Entity.Pos.X, Y: s.Entity.Pos.Y + 0.5, Z: s.Entity.Pos.Z}.BlockPos()
		return world.Vec{X: float64(p.X), Y: float64(p.Y), Z: float64(p.Z)}
	}
	b := s.Pos.BlockPos()
	return world.Vec{X: float64(b.X), Y: float64(b.Y), Z: float64(b.Z)}
}

// WithEntity returns a copy of the source executed by entity.
//
// The listener is no longer the player: the command of an `/execute` runs
// as the target entity, but the listener is the wrapper of the execute.
func (s *Source) WithEntity(e *world.Entity) *Source {
	cp := s.Clone()
	cp.Entity = e
	cp.ListenerPlayer = false
	return cp
}

// WithPermissionLevel returns a copy of the source with the given
// permission level.
//
// 1.12.2 caps the permission level of a listener with
// `CommandListenerWrapper.a(int)`, which keeps the *lower* of the two
// levels. A function therefore runs its commands at level 2, even when
// the caller is the console.
func (s *Source) WithPermissionLevel(level int) *Source {
	cp := s.Clone()
	cp.PermissionLevel = level
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
	// permissions holds the minimum permission level a command needs.
	permissions map[string]int
	// extensions holds the commands gomc adds on top of the 1.12.2 set,
	// with the note that says what each of them is for.
	extensions map[string]string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers:    map[string]Handler{},
		permissions: map[string]int{},
		extensions:  map[string]string{},
	}
}

// Register adds a command handler.
func (r *Registry) Register(h Handler) {
	if _, ok := r.handlers[h.Name()]; !ok {
		r.order = append(r.order, h.Name())
	}
	r.handlers[h.Name()] = h
}

// SetExtension marks a command as one gomc adds on top of the 1.12.2 set.
//
// 1.12.2 refuses a function that uses a command it does not know, so a
// function that uses one of these loads on gomc but not on a real server.
// The check reports the difference, so that it is never silent.
func (r *Registry) SetExtension(name, note string) {
	if r.extensions == nil {
		r.extensions = map[string]string{}
	}
	r.extensions[name] = note
}

// Extension returns the note of an extension command.
func (r *Registry) Extension(name string) (string, bool) {
	note, ok := r.extensions[name]
	return note, ok
}

// ExtensionNames returns the names of the extension commands, sorted.
func (r *Registry) ExtensionNames() []string {
	out := make([]string, 0, len(r.extensions))
	for name := range r.extensions {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// SetRequiredPermission declares the minimum permission level a command
// needs.
//
// 1.12.2 asks a command whether a source may run it (`ICommand.canUse`),
// and the command answers with the level it declares. A command without a
// declared level can be run by anybody.
func (r *Registry) SetRequiredPermission(name string, level int) {
	if r.permissions == nil {
		r.permissions = map[string]int{}
	}
	r.permissions[name] = level
}

// RequiredPermission returns the minimum permission level the command
// needs, or 0 when it has none.
func (r *Registry) RequiredPermission(name string) int {
	return r.permissions[name]
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
//
// It returns nil when the argument is not an NBT one or when the list is
// too short: the position is a convention between the command and the
// parser, and a wrong one must not take the whole run down.
func NBTOf(args []ast.Arg, i int) *nbt.Compound {
	if i < 0 || i >= len(args) {
		return nil
	}
	if a, ok := args[i].(*ast.NBTArg); ok {
		return a.Value
	}
	return nil
}

// SelectorOf is a helper to read a selector argument from a parsed list.
//
// It returns nil when the argument is not a selector one or when the list
// is too short, see NBTOf.
func SelectorOf(args []ast.Arg, i int) *ast.Selector {
	if i < 0 || i >= len(args) {
		return nil
	}
	if a, ok := args[i].(*ast.Selector); ok {
		return a
	}
	return nil
}
