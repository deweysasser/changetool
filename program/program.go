package program

import (
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// Version is created by the Makefile and passed in as a linker flag.  When go 1.18 is released, this will be replaced
// with the built-in mechanism

var Version = "unknown"

// Options is the structure of program options
type Options struct {
	Debug      bool       `short:"d" help:"Show debugging information"`
	LogFormat  string     `short:"l" enum:"auto,jsonl,terminal" default:"auto" help:"How to show program output (auto|terminal|jsonl)"`
	Quiet      bool       `short:"q" help:"Be less verbose than usual"`
	Path       string     `default:"." type:"existingdir" short:"p" help:"Path for the git worktree/repo to log"`
	Output     string     `default:"-" short:"o" help:"File to which to send output"`
	OutFP      *os.File   `kong:"-"`
	Changelog  Changelog  `cmd:""`
	VersionCmd VersionCmd `name:"version" cmd:"" help:"show program version"`
	Semver     Semver     `cmd:"" help:"Manipulate Semantic Versions"`
}

// Run runs the program
func (program *Options) Run(_ *Options) error {
	return nil
}

func (program *Options) AfterApply() error {
	program.Init()

	if program.Output == "-" {
		program.OutFP = os.Stdout
		return nil
	} else {
		log.Debug().Str("file", program.Output).Msg("Sending output")
		fp, err := os.Create(program.Output)
		if err != nil {
			return err
		}

		program.OutFP = fp
		return nil
	}
}

// Repository opens the git repository
func (program *Options) Repository() (*git.Repository, error) {
	log.Debug().Str("path", program.Path).Msg("Opening repository")
	return git.PlainOpen(program.Path)
}

func (program *Options) Init() {

	switch {
	case program.Debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case program.Quiet:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if program.LogFormat == "terminal" ||
		(program.LogFormat == "auto" && isTerminal(os.Stdout)) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		log.Logger = log.Output(os.Stdout)
	}

	log.Logger.Debug().
		Str("version", Version).
		Str("program", os.Args[0]).
		Msg("Starting")
}

// isTerminal returns true if the file given points to a character device (i.e. a terminal)
func isTerminal(file *os.File) bool {
	if fileInfo, err := file.Stat(); err != nil {
		log.Err(err).Msg("Error running stat")
		return false
	} else {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
}
