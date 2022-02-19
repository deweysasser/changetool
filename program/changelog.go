package program

import (
	"fmt"
	"github.com/deweysasser/changetool/changes"
	"github.com/deweysasser/changetool/perf"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"strings"
)

type Changelog struct {
	MaxCommits             int               `short:"n" group:"source" default:"1000" help:"max number of commits to check"`
	SinceTag               string            `short:"s" group:"source" help:"Tag from which to start" aliases:"since"`
	AllCommits             bool              `short:"a" group:"source" help:"report changelog on all commits up to --max-commits.  Otherwise, report only to last version tag"`
	DefaultType            changes.TypeTag   `default:"fix" group:"calculation" help:"if type is not specified in commit, assume this type"`
	GuessMissingCommitType bool              `default:"true" group:"calculation" negatable:"" help:"If commit type is missing, take a guess about which it is"`
	Order                  []changes.TypeTag `default:"${type_order}" group:"calculation" help:"order in which to list commit message types"`
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

	var guess changes.CommitTypeGuesser
	if c.GuessMissingCommitType {
		guess = c.guessType
	} else {
		guess = func(commit *object.Commit) changes.TypeTag {
			return c.DefaultType
		}
	}

	if stopAt, err := c.findStopCriteria(r); err != nil {
		return nil, err
	} else {
		return changes.Load(r, stopAt, guess)
	}
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

func (c *Changelog) findStopCriteria(r *repo.Repository) (stop changes.StopAt, err error) {
	switch {
	case c.SinceTag != "":
		if hash, found := r.TagMap()[c.SinceTag]; !found {
			return nil, fmt.Errorf("unable to find start tag %s", c.SinceTag)
		} else {
			log.Debug().Str("tag", c.SinceTag).Msg("Stopping at tag")
			return changes.StopAtHash(hash), nil
		}
	case c.AllCommits:
		log.Debug().Int("count", c.MaxCommits).Msg("stopping after # of commits")
		return changes.StopAtCount(c.MaxCommits), nil
	default:
		log.Debug().
			Msg("stopping at first recognizable semver")
		return changes.StopAtFirstSemver(r), nil
	}
}
