package changes

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

func (c *ChangeSet) addCommit(tt TypeTag, section string, message string) {
	c.Commits[tt] = append(c.Commits[tt], message)
}

// CommitTypeGuesser is a guess function to guess commit type from the commit.  StandardGuess can be used as a base to fill this in.
type CommitTypeGuesser func(commit *object.Commit) TypeTag

var commitType = regexp.MustCompile(`([a-zA-Z_][a-zA-Z_0-9]*)(\(([a-zA-Z_][a-zA-Z_0-9]*)\))?(!)?: *`)

// Load creates a new CommitSet from a repository
func Load(r *git.Repository, stopAt plumbing.Hash, guess CommitTypeGuesser) (*ChangeSet, error) {

	changeSet := NewChangeSet()

	log.Debug().
		Str("stop_at", stopAt.String()[:6]).
		Msg("Stopping at commit")
	iter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})

	if err != nil {
		return nil, err
	}

	defer iter.Close()

	numChanges := 0

	_ = iter.ForEach(func(commit *object.Commit) error {
		if len(commit.ParentHashes) > 1 {
			return nil
		}

		log.Debug().
			Str("stop_at", stopAt.String()[:6]).
			Str("this_commit", commit.Hash.String()[:6]).
			Msg("Looking for stop point")

		if commit.Hash == stopAt {
			return storer.ErrStop
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
