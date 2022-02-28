package program

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/deweysasser/changetool/changes"
	"github.com/deweysasser/changetool/repo"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"runtime"
)

// Options is the structure of program options
type Options struct {
	Debug     bool   `short:"d" group:"Info" help:"Show debugging information"`
	LogFormat string `short:"l" group:"Info" enum:"auto,jsonl,terminal" default:"auto" help:"How to show program output (auto|terminal|jsonl)"`
	Quiet     bool   `short:"q" group:"Info" help:"Be less verbose than usual"`

	Path   string `default:"." group:"locations" type:"existingdir" short:"p" help:"Path for the git worktree/repo to log"`
	Output string `default:"-" group:"locations" short:"o" help:"File to which to send output"`

	Changelog  Changelog  `cmd:"" help:"calculate changelogs"`
	VersionCmd VersionCmd `name:"version" cmd:"" help:"show program version"`
	Semver     Semver     `cmd:"" help:"Manipulate Semantic Versions"`

	OutFP *os.File `kong:"-"`
}

// Parse calls the CLI parsing routines
func (program *Options) Parse(args []string) (*kong.Context, error) {
	parser, err := kong.New(program,
		kong.Description("Brief Program Summary"),
		kong.ShortUsageOnError(),
		kong.Vars{
			"type_order": changes.TypesInOrder.Join(","),
		},
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return parser.Parse(args)
}

// Run runs the program
func (program *Options) Run(_ *Options) error {
	return nil
}

// AfterApply runs after the options are parsed but before anything runs
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
func (program *Options) Repository() (*repo.Repository, error) {
	log.Debug().Str("path", program.Path).Msg("Opening repository")
	return repo.New(program.Path)
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

	var out io.Writer = os.Stdout

	if os.Getenv("TERM") == "" && runtime.GOOS == "windows" {
		out = colorable.NewColorableStdout()
	}

	if program.LogFormat == "terminal" ||
		(program.LogFormat == "auto" && isTerminal(os.Stdout)) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: out})
	} else {
		log.Logger = log.Output(out)
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
