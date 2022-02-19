package changes

import (
	"github.com/deweysasser/changetool/perf"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

// ChangeSet represents the list of all changes in a commit stream (presumably from a base tag)
type ChangeSet struct {
	BreakingChanges []string
	Commits         map[TypeTag][]string
}

// NewChangeSet creates a new, empty change set
func NewChangeSet() *ChangeSet {
	return &ChangeSet{Commits: make(map[TypeTag][]string)}
}

func (c *ChangeSet) addBreaking(message string) {
	c.BreakingChanges = append(c.BreakingChanges, message)
}

// TODO:  use the section (middle argument)
func (c *ChangeSet) addCommit(tt TypeTag, _ string, message string) {
	c.Commits[tt] = append(c.Commits[tt], message)
}

// CommitTypeGuesser is a guess function to guess commit type from the commit.  StandardGuess can be used as a base to fill this in.
type CommitTypeGuesser func(commit *object.Commit) TypeTag

// DefaultGuess just returns the specified tag if it has to guess
func DefaultGuess(tag TypeTag) CommitTypeGuesser {
	return func(commit *object.Commit) TypeTag {
		return tag
	}
}

var commitType = regexp.MustCompile(`([a-zA-Z_][a-zA-Z_0-9]*)(\(([a-zA-Z_][a-zA-Z_0-9]*)\))?(!)?: *`)

// Load creates a new CommitSet from a repository
func Load(r *repo.Repository, stopAt StopAt, guess CommitTypeGuesser) (*ChangeSet, error) {
	defer perf.Timer("Loading changes").Stop()

	changeSet := NewChangeSet()

	iter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})

	if err != nil {
		return nil, err
	}

	defer iter.Close()

	numChanges := 0
	_ = iter.ForEach(func(commit *object.Commit) error {

		log.Debug().
			Str("this_commit", commit.Hash.String()[:6]).
			Msg("Examining Commit")

		if stopAt(commit) {
			return storer.ErrStop
		}

		if len(commit.ParentHashes) > 1 {
			return nil
		}

		numChanges++
		message := commit.Message
		re := commitType.FindStringSubmatch(message)
		section := ""
		if len(re) > 3 {
			section = re[3]
		}

		var tt TypeTag

		if len(re) > 1 {
			tt = TypeTag(re[1])
			message = message[len(re[0]):]
		} else {
			tt = guess(commit)
		}

		changeSet.addCommit(tt, section, message)
		if (len(re) > 4 && re[4] != "") || strings.Contains(message, "BREAKING CHANGE") {
			changeSet.addBreaking(message)
		}

		return nil
	})

	log.Debug().
		Int("number_of_changes", numChanges).
		Msg("Number of changes")

	return changeSet, nil
}
