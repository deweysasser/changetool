package program

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
)

// Version is created by the Makefile and passed in as a linker flag
var Version = "unknown"

// Options is the structure of program options
type Options struct {
	Debug   bool `short:"d" help:"Show debugging information"`
	Version bool `short:"v" help:"Show program version"`
	Quiet   bool `short:"q" help:"Be less verbose than usual"`
}

// Run runs the program
func (program *Options) Run() error {

	if program.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	if program.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return nil
}
