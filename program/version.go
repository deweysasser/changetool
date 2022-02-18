package program

import "fmt"

type VersionCmd struct{}

func (v *VersionCmd) Run(program *Options) error {
	_, _ = fmt.Fprintln(program.OutFP, Version)
	return nil
}
