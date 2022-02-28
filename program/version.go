package program

import "fmt"

// Version is created by the Makefile and passed in as a linker flag.  When go 1.18 is released, this will be replaced
// with the built-in mechanism

var Version = "unknown"

// VersionCmd prints the program version
type VersionCmd struct{}

func (v *VersionCmd) Run(program *Options) error {
	_, _ = fmt.Fprintln(program.OutFP, Version)

	return nil
}
