package program

import (
	"errors"
	"fmt"
	"github.com/deweysasser/changetool/changes"
	"github.com/deweysasser/changetool/perf"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"strings"
)

type Changelog struct {
	MaxCommits             int               `short:"n" default:"1000" help:"max number of commits to check"`
	SinceTag               string            `short:"s" help:"Tag from which to start" aliases:"since"`
	DefaultType            changes.TypeTag   `default:"fix" help:"if type is not specified in commit, assume this type"`
	GuessMissingCommitType bool              `default:"true" negatable:"" help:"If commit type is missing, take a guess about which it is"`
	Order                  []changes.TypeTag `default:"${type_order}" help:"order in which to list commit message types"`
}

func (c *Changelog) Run(program *Options) error {

	r, err := program.Repository()
	if err != nil {
		return err
	}

	changeSet, err := c.CalculateChanges(r)
	if err != nil {
		return err
	}

	for _, section := range changes.CommitEntries(c.Order, changeSet.Commits) {
		_, _ = fmt.Fprintf(program.OutFP, "%s:\n", section.Name)

		for _, commit := range section.Messages {
			message := commit
			if message[len(message)-1] == '\n' {
				message = message[:len(message)-1]
			}

			_, _ = fmt.Fprintf(program.OutFP, "   * %s", strings.ReplaceAll(message, "\n", "\n     "))
			_, _ = fmt.Fprintln(program.OutFP)

		}
		_, _ = fmt.Fprintln(program.OutFP)
	}

	return nil
}

func (c *Changelog) CalculateChanges(r *repo.Repository) (*changes.ChangeSet, error) {
	defer perf.Timer("Calculating Changes").Stop()

	tagsTimer := perf.Timer("reading tags")

	stopAt, err := c.findStartVersion(r)
	if err != nil {
		return nil, err
	}

	tagsTimer.Stop()

	var guess changes.CommitTypeGuesser
	if c.GuessMissingCommitType {
		guess = c.guessType
	} else {
		guess = func(commit *object.Commit) changes.TypeTag {
			return c.DefaultType
		}
	}

	return changes.Load(r, stopAt, guess, c.MaxCommits)
}

// guessType guesses the type of the commit from information in the commit
func (c *Changelog) guessType(commit *object.Commit) changes.TypeTag {
	tag, err := changes.StandardGuess(commit)
	if err != nil {
		log.Debug().AnErr("err", err).
			Str("hash", commit.Hash.String()[:6]).
			Msg("No guess for commit")
		return c.DefaultType
	}
	return tag
}

func (c *Changelog) findStartVersion(r *repo.Repository) (stop plumbing.Hash, err error) {
	defer perf.Timer("finding start version")
	if c.SinceTag != "" {
		if hash, found := r.TagMap()[c.SinceTag]; !found {
			return plumbing.ZeroHash, errors.New("unable to find desired tag " + c.SinceTag)
		} else {
			return hash, nil
		}
	}
	return plumbing.ZeroHash, nil

}
