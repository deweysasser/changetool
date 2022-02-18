package program

import (
	"errors"
	"fmt"
	"github.com/deweysasser/changetool/changes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"strings"
)

type Changelog struct {
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

func (c *Changelog) CalculateChanges(r *git.Repository) (*changes.ChangeSet, error) {

	tags, err := r.Tags()
	if err != nil {
		return nil, err
	}

	defer tags.Close()

	stopAt, err := c.findStartVersion(tags)
	if err != nil {
		return nil, err
	}

	var guess changes.CommitTypeGuesser
	if c.GuessMissingCommitType {
		guess = c.guessType
	} else {
		guess = func(commit *object.Commit) changes.TypeTag {
			return c.DefaultType
		}
	}

	return changes.Load(r, stopAt, guess)
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

func (c *Changelog) findStartVersion(tags storer.ReferenceIter) (stop plumbing.Hash, err error) {
	if c.SinceTag != "" {
		lookFor := c.SinceTag

		log.Debug().Str("tag", c.SinceTag).Msg("Looking for tag")
		_ = tags.ForEach(func(t *plumbing.Reference) error {
			name := t.Name().Short()
			log.Debug().
				Str("tag", name).Msg("comparing tag")
			if name == lookFor {
				log.Debug().
					Str("tag", c.SinceTag).
					Str("hash", t.Hash().String()).
					Msg("Found hash for tag")

				stop = t.Hash()
				return storer.ErrStop
			}

			return nil
		})

		if stop == plumbing.ZeroHash {

			err = errors.New("unable to find desired tag " + c.SinceTag)
			return
		}
	}

	return
}
