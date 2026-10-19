// Package gomc is the command line entry point of the Minecraft 1.12.2
// mcfunction runtime.
//
// The real work lives in internal/cli, which dispatches the subcommands:
// check, run, tick, commands and difftest.
package main

import (
	"os"

	"github.com/mcsim/gomc/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
