package program

import "fmt"

type VersionCmd struct{}

func (v *VersionCmd) Run() error {
	fmt.Println(Version)
	return nil
}
