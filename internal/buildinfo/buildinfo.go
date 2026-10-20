// Package buildinfo carries the version of the gomc build.
//
// The values are overridden at build time, so a released binary can say
// exactly which source it was built from:
//
//	go build -ldflags "\
//	  -X github.com/Mcd0LUO/gomc/internal/buildinfo.Version=v0.1.0 \
//	  -X github.com/Mcd0LUO/gomc/internal/buildinfo.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/Mcd0LUO/gomc/internal/buildinfo.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	  -o gomc.exe ./cmd/gomc
//
// The defaults describe a build straight from the working tree.
package buildinfo

// Version is the gomc version.
//
// It follows semantic versioning: the first release of a tool whose
// behaviour is still being aligned with a moving target is 0.x.y.
//
// It is a variable, not a constant, because the linker can only replace
// variables: `-X` on a constant is silently ignored.
var Version = "0.1.0"

// Commit is the git commit the binary was built from, when the build
// system passed one in.
var Commit = "unknown"

// Date is the build time in UTC, when the build system passed one in.
var Date = "unknown"

// String renders the version for `gomc --version`.
func String() string {
	s := "gomc " + Version
	if Commit != "" && Commit != "unknown" {
		s += " (" + Commit
		if Date != "" && Date != "unknown" {
			s += ", built " + Date
		}
		s += ")"
	}
	return s
}
