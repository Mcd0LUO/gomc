package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Mcd0LUO/gomc/internal/exec"
)

// The bench subcommand measures a project.
//
// It loads the functions of a world the way the server does, runs them,
// and reports what happened: how many commands there were, which commands
// were used, how long each function took, and what failed.
//
// The numbers come from the runtime itself, so they are the numbers of
// **this** runtime -- they say how the project behaves under gomc, which
// is what the person who wrote it can act on.

// benchOptions are the settings of a measurement.
type benchOptions struct {
	// function is the function to run. It is empty when every function
	// is run.
	function string
	// ticks is the number of ticks to run. It is -1 when the ticks
	// are not used.
	ticks int
	// top is the number of rows the detailed report has.
	top int
	// json prints the report as JSON.
	json bool
	// as is the player the functions run as, empty for the console.
	as string
	// repeat is how many times the run is repeated.
	//
	// A single command is quicker than the clock can tell, so the
	// numbers of one run are mostly zero. Repeating the run adds the
	// times up until they say something.
	repeat int
	// all says to run every function, one after the other.
	all bool
}

// benchReport is the result of a measurement.
//
// The shape is the one the JSON output has, so a script and a person
// read the same numbers.
type benchReport struct {
	// World is the directory that was measured.
	World string `json:"world"`
	// Mode says what was run.
	Mode string `json:"mode"`
	// Functions is how many functions the world has.
	Functions int `json:"functions"`
	// Commands is how many commands ran.
	Commands int `json:"commands"`
	// Failed is how many of them failed.
	Failed int `json:"failed"`
	// Errors is how many distinct failures there were.
	Errors int `json:"errors"`
	// Duration is how long the run took.
	Duration string `json:"duration"`
	// LoadDuration is how long the loading of the functions took.
	// It is a part of the duration above, and it is worth knowing
	// on its own: a project with many functions is slow to load.
	LoadDuration string `json:"load_duration"`
	// LoadDurationNS is the same time in nanoseconds.
	LoadDurationNS int64 `json:"load_duration_ns"`
	// DurationNS is the same time in nanoseconds, for the scripts.
	DurationNS int64 `json:"duration_ns"`
	// ByCommand groups the commands by their name.
	ByCommand []benchCommandGroup `json:"by_command"`
	// ByFunction groups the commands by the function they ran in.
	ByFunction []benchFunctionGroup `json:"by_function"`
}

// benchCommandGroup is the measurement of one command.
type benchCommandGroup struct {
	// Command is the name of the command.
	Command string `json:"command"`
	// Count is how many times it ran.
	Count int `json:"count"`
	// Failed is how many of them failed.
	Failed int `json:"failed"`
	// Duration is the time of all of them.
	Duration string `json:"duration"`
	// DurationNS is the same time in nanoseconds.
	DurationNS int64 `json:"duration_ns"`
}

// benchFunctionGroup is the measurement of one function.
type benchFunctionGroup struct {
	// Function is the identifier of the function.
	Function string `json:"function"`
	// Count is how many commands ran in it.
	Count int `json:"count"`
	// Failed is how many of them failed.
	Failed int `json:"failed"`
	// Duration is the time of all of them.
	Duration string `json:"duration"`
	// DurationNS is the same time in nanoseconds.
	DurationNS int64 `json:"duration_ns"`
}

// runBench measures the functions of a world.
func runBench(args []string, stdout, stderr io.Writer, opts options) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "bench requires a world directory")
		return 2
	}
	world := args[0]
	bopts, err := parseBenchOptions(args[1:], opts)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	report, err := measure(world, bopts, opts.as, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if bopts.json {
		return printBenchJSON(stdout, report)
	}
	printBenchReport(stdout, report, bopts)
	return 0
}

// parseBenchOptions reads the switches of the bench subcommand.
func parseBenchOptions(args []string, opts options) (benchOptions, error) {
	bopts := benchOptions{ticks: -1, top: 20, repeat: 1, as: opts.as, json: opts.json}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--function", "-f":
			if i+1 >= len(args) {
				return bopts, fmt.Errorf("--function needs the id of a function")
			}
			i++
			bopts.function = args[i]
		case "--ticks", "-t":
			if i+1 >= len(args) {
				return bopts, fmt.Errorf("--ticks needs a number")
			}
			i++
			n, err := parseInt(args[i])
			if err != nil {
				return bopts, fmt.Errorf("--ticks: %w", err)
			}
			bopts.ticks = n
		case "--top", "-n":
			if i+1 >= len(args) {
				return bopts, fmt.Errorf("--top needs a number")
			}
			i++
			n, err := parseInt(args[i])
			if err != nil {
				return bopts, fmt.Errorf("--top: %w", err)
			}
			bopts.top = n
		case "--repeat", "-r":
			if i+1 >= len(args) {
				return bopts, fmt.Errorf("--repeat needs a number")
			}
			i++
			n, err := parseInt(args[i])
			if err != nil {
				return bopts, fmt.Errorf("--repeat: %w", err)
			}
			if n < 1 {
				return bopts, fmt.Errorf("--repeat needs a number of at least 1")
			}
			bopts.repeat = n
		case "--all", "-a":
			bopts.all = true
		case "--json":
			bopts.json = true
		default:
			return bopts, fmt.Errorf("unknown bench switch %q", args[i])
		}
	}
	return bopts, nil
}

// benchSource returns the source the functions run as.
//
// The console is the default, and a player can be asked for with --as.
func benchSource(engine *exec.Engine, player string) (*exec.Source, error) {
	if player != "" {
		return playerSource(engine, player)
	}
	return &exec.Source{
		Name:            "Server",
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}, nil
}

// measure runs what the options ask for and collects the numbers.
func measure(world string, bopts benchOptions, as string, stderr io.Writer) (*benchReport, error) {
	report := &benchReport{World: world}
	start := time.Now()
	var loadDuration time.Duration
	for rep := 0; rep < bopts.repeat; rep++ {
		// Every repetition gets a **fresh world**, so that the runs are
		// comparable: a function that adds an objective would fail on
		// the second run, because the objective is already there.
		engine, err := newEngine(world)
		if err != nil {
			return nil, err
		}
		loadStart := time.Now()
		loadErrs := engine.LoadFunctions()
		loadDuration += time.Since(loadStart)
		if len(loadErrs) > 0 && rep == 0 {
			// The broken functions are reported by the check, but the
			// measurement still goes on with the ones that loaded. The
			// warning is printed once, not once per repetition.
			fmt.Fprintf(stderr, "warning: %d function(s) failed to load\n", len(loadErrs))
		}
		stats := &exec.Stats{}
		engine.SetCommandRecorder(stats)
		src, err := benchSource(engine, as)
		if err != nil {
			return nil, err
		}
		mode, err := runMeasured(engine, bopts, src)
		if err != nil {
			return nil, err
		}
		report.Mode = mode
		report.Functions = len(engine.Functions.IDs())
		report.Commands += stats.Count()
		report.Errors += len(engine.Errors)
		for _, g := range stats.ByCommand() {
			report.Failed += g.Failed
			report.ByCommand = mergeCommandGroups(report.ByCommand, g)
		}
		for _, g := range stats.ByFunction() {
			report.ByFunction = mergeFunctionGroups(report.ByFunction, g)
		}
	}
	if bopts.repeat > 1 {
		report.Mode = report.Mode + fmt.Sprintf(", %d times", bopts.repeat)
	}
	elapsed := time.Since(start)
	report.Duration = elapsed.String()
	report.DurationNS = elapsed.Nanoseconds()
	report.LoadDuration = formatBenchTime(loadDuration)
	report.LoadDurationNS = loadDuration.Nanoseconds()
	return report, nil
}

// runMeasured runs what the options ask for, once.
//
// It returns what it ran, so that the report can name the mode.
func runMeasured(engine *exec.Engine, bopts benchOptions, src *exec.Source) (string, error) {
	switch {
	case bopts.function != "":
		_, err := engine.RunFunctionByID(bopts.function, src)
		return "function " + bopts.function, err
	case bopts.ticks >= 0:
		for i := 0; i < bopts.ticks; i++ {
			if _, err := engine.Tick(src); err != nil {
				return "", err
			}
		}
		return fmt.Sprintf("%d tick(s)", bopts.ticks), nil
	case bopts.all:
		for _, id := range engine.Functions.IDs() {
			if _, err := engine.RunFunctionByID(id, src); err != nil {
				return "", err
			}
		}
		return "every function", nil
	default:
		// The default is the game loop function, because that is the
		// one a server runs on its own.
		id := engine.World.Rules.String("gameLoopFunction")
		if id == "" || id == "-" {
			return "", fmt.Errorf("no gameLoopFunction game rule is set, so there is nothing to run; use --all, --function or --ticks")
		}
		for i := 0; i < 10; i++ {
			if _, err := engine.Tick(src); err != nil {
				return "", err
			}
		}
		return "gameLoopFunction " + id, nil
	}
}

// mergeCommandGroups adds a group to the list of groups, adding the
// numbers of the two together.
func mergeCommandGroups(list []benchCommandGroup, g exec.CommandGroup) []benchCommandGroup {
	for i := range list {
		if list[i].Command == g.Command {
			list[i].Count += g.Count
			list[i].Failed += g.Failed
			list[i].DurationNS += g.Duration.Nanoseconds()
			list[i].Duration = formatBenchTime(time.Duration(list[i].DurationNS))
			return list
		}
	}
	return append(list, benchCommandGroup{
		Command:    g.Command,
		Count:      g.Count,
		Failed:     g.Failed,
		DurationNS: g.Duration.Nanoseconds(),
		Duration:   formatBenchTime(g.Duration),
	})
}

// mergeFunctionGroups adds a group to the list of groups, adding the
// numbers of the two together.
func mergeFunctionGroups(list []benchFunctionGroup, g exec.FunctionGroup) []benchFunctionGroup {
	for i := range list {
		if list[i].Function == g.Function {
			list[i].Count += g.Count
			list[i].Failed += g.Failed
			list[i].DurationNS += g.Duration.Nanoseconds()
			list[i].Duration = formatBenchTime(time.Duration(list[i].DurationNS))
			return list
		}
	}
	return append(list, benchFunctionGroup{
		Function:   g.Function,
		Count:      g.Count,
		Failed:     g.Failed,
		DurationNS: g.Duration.Nanoseconds(),
		Duration:   formatBenchTime(g.Duration),
	})
}

// formatBenchTime renders a duration the way a person reads it.
//
// A time that is shorter than a microsecond is shown in nanoseconds,
func formatBenchTime(d time.Duration) string {
	switch {
	case d == 0:
		return "0"
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.1fus", float64(d.Nanoseconds())/1000)
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000)
	default:
		return d.String()
	}
}

// printBenchReport prints the measurement for a person to read.
func printBenchReport(w io.Writer, report *benchReport, bopts benchOptions) {
	fmt.Fprintf(w, "world:  %s\n", report.World)
	fmt.Fprintf(w, "mode:   %s\n", report.Mode)
	fmt.Fprintf(w, "result: %d function(s), %d command(s), %d failure(s), %d error(s), %s\n",
		report.Functions, report.Commands, report.Failed, report.Errors, report.Duration)
	fmt.Fprintf(w, "load:   %s of the time above was loading the functions\n", report.LoadDuration)

	if len(report.ByFunction) > 0 {
		fmt.Fprintf(w, "\nby function (the %d slowest of %d):\n", min(bopts.top, len(report.ByFunction)), len(report.ByFunction))
		fmt.Fprintf(w, "  %-46s %8s %10s %8s\n", "function", "commands", "time", "failed")
		for i, g := range report.ByFunction {
			if i >= bopts.top {
				break
			}
			fmt.Fprintf(w, "  %-46s %8d %10s %8d\n", g.Function, g.Count, g.Duration, g.Failed)
		}
	}
	if len(report.ByCommand) > 0 {
		fmt.Fprintf(w, "\nby command:\n")
		fmt.Fprintf(w, "  %-24s %8s %10s %8s\n", "command", "count", "time", "failed")
		for _, g := range report.ByCommand {
			fmt.Fprintf(w, "  %-24s %8d %10s %8d\n", g.Command, g.Count, g.Duration, g.Failed)
		}
	}
}

// printBenchJSON prints the measurement for a script to read.
func printBenchJSON(w io.Writer, report *benchReport) int {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		// The report could not be written, so the error is reported in
		// the same shape. The text is marshalled, so no escape
		// sequence is needed here.
		fallback := map[string]string{"error": err.Error()}
		if data, merr := json.Marshal(fallback); merr == nil {
			fmt.Fprintln(w, string(data))
		}
		return 1
	}
	return 0
}

// min returns the smaller of the two numbers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
