package exec

import (
	"sort"
	"time"
)

// CommandStat is what a measurement of one command says.
//
// The engine reports every command it runs, so a caller can measure a
// whole project: how many commands there are, which ones are used, and
// where the time goes.
type CommandStat struct {
	// Command is the name of the command, e.g. "setblock".
	Command string
	// Function is the identifier of the function the command ran in.
	//
	// It is empty for a command that did not run inside a function.
	Function string
	// Duration is how long the command took.
	Duration time.Duration
	// Failed reports whether the command failed.
	Failed bool
}

// CommandRecorder is told about every command the engine runs.
//
// The recorder is a measuring tool, so it lives behind an interface: the
// engine knows nothing about the reports, and a version package or a tool
// decides what to do with them.
type CommandRecorder interface {
	// Record is told about a command that ran.
	//
	// The recorder is called after the command finished, so the
	// duration covers the whole command.
	Record(stat CommandStat)
}

// SetCommandRecorder sets the recorder that is told about every command.
//
// An engine without a recorder measures nothing.
func (e *Engine) SetCommandRecorder(rec CommandRecorder) {
	e.recorder = rec
}

// CommandRecorder returns the recorder of the engine, if it has one.
func (e *Engine) CommandRecorder() CommandRecorder {
	return e.recorder
}

// Stats collects the measurements of a run.
//
// It is a recorder, so it can be given to the engine directly. It is
// safe to use one collector for many runs: the counts add up.
type Stats struct {
	// stats holds the measurements, in the order they arrived.
	stats []CommandStat
	// total is the time of all the commands.
	total time.Duration
}

// Record implements CommandRecorder.
func (s *Stats) Record(stat CommandStat) {
	s.stats = append(s.stats, stat)
	s.total += stat.Duration
}

// Total returns the time of all the commands that ran.
func (s *Stats) Total() time.Duration { return s.total }

// Count returns how many commands ran.
func (s *Stats) Count() int { return len(s.stats) }

// All returns the measurements, in the order they arrived.
func (s *Stats) All() []CommandStat { return s.stats }

// ByCommand groups the measurements by the name of the command.
//
// The groups come back **slowest first**, and the names that tie are
// sorted alphabetically, so the report is stable.
func (s *Stats) ByCommand() []CommandGroup {
	groups := map[string]*CommandGroup{}
	for _, st := range s.stats {
		g, ok := groups[st.Command]
		if !ok {
			g = &CommandGroup{Command: st.Command}
			groups[st.Command] = g
		}
		g.Count++
		g.Duration += st.Duration
		if st.Failed {
			g.Failed++
		}
	}
	out := make([]CommandGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Duration != out[j].Duration {
			return out[i].Duration > out[j].Duration
		}
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Command < out[j].Command
	})
	return out
}

// ByFunction groups the measurements by the function they ran in.
//
// The functions come back **slowest first**, because that is the order a
// person wants them in: the ones that cost the most are the ones worth
// looking at. The functions that tie are sorted by their identifier, so
// the report is stable.
func (s *Stats) ByFunction() []FunctionGroup {
	groups := map[string]*FunctionGroup{}
	for _, st := range s.stats {
		g, ok := groups[st.Function]
		if !ok {
			g = &FunctionGroup{Function: st.Function}
			groups[st.Function] = g
		}
		g.Count++
		g.Duration += st.Duration
		if st.Failed {
			g.Failed++
		}
	}
	out := make([]FunctionGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Duration != out[j].Duration {
			return out[i].Duration > out[j].Duration
		}
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Function < out[j].Function
	})
	return out
}

// CommandGroup is the measurement of one command.
type CommandGroup struct {
	// Command is the name of the command.
	Command string
	// Count is how many times it ran.
	Count int
	// Duration is the time of all of them.
	Duration time.Duration
	// Failed is how many of them failed.
	Failed int
}

// FunctionGroup is the measurement of one function.
type FunctionGroup struct {
	// Function is the identifier of the function.
	Function string
	// Count is how many commands ran in it.
	Count int
	// Duration is the time of all of them.
	Duration time.Duration
	// Failed is how many of them failed.
	Failed int
}
