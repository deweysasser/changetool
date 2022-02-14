package program

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"path"
	"regexp"
	"strings"
)

type Changelog struct {
	Path                   string    `default:"." short:"p" help:"Path for the git worktree/repo to log"`
	SinceTag               string    `short:"s" help:"Tag from which to start" aliases:"since"`
	DefaultType            TypeTag   `default:"fix" help:"if type is not specified in commit, assume this type"`
	GuessMissingCommitType bool      `default:"true" negatable:"" help:"If commit type is missing, take a guess about which it is"`
	Order                  []TypeTag `default:"${type_order}" help:"order in which to list commit message types"`
}

var commitType = regexp.MustCompile(`([a-zA-Z_][a-zA-Z_0-9]*)(\(([a-zA-Z_][a-zA-Z_0-9]*)\))?(!)?: *`)

func (c *Changelog) Run() error {

	changeSet, err := c.CalculateChanges()
	if err != nil {
		return err
	}

	for _, section := range asCommitList(c.Order, changeSet.Commits) {
		fmt.Printf("%s:\n", section.Name)

		for _, commit := range section.Messages {
			message := commit
			if message[len(message)-1] == '\n' {
				message = message[:len(message)-1]
			}

			fmt.Printf("   * %s", strings.ReplaceAll(message, "\n", "\n     "))
			fmt.Println()

		}
		fmt.Println()
	}

	return nil
}

func (c *Changelog) CalculateChanges() (*ChangeSet, error) {
	changeSet := NewChangeset()
	r, err := git.PlainOpen(c.Path)
	if err != nil {
		return nil, err
	}

	tags, err := r.Tags()
	if err != nil {
		return nil, err
	}

	defer tags.Close()

	stopAt, err := c.findStartVersion(tags)
	if err != nil {
		return nil, err
	}

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

		switch {
		case len(re) > 1:
			tt = TypeTag(re[1])
			message = message[len(re[0]):]
		case c.GuessMissingCommitType:
			tt = c.guessType(commit)
		default:
			tt = c.DefaultType
		}

		changeSet.AddCommit(tt, section, message)
		if (len(re) > 4 && re[4] != "") || strings.Contains(message, "BREAKING CHANGE") {
			changeSet.AddBreaking(message)
		}
		return nil
	})

	log.Debug().
		Int("number_of_changes", numChanges).
		Msg("Number of changes")

	return changeSet, nil
}

// guessType guesses the type of the commit from information in the commit
func (c *Changelog) guessType(commit *object.Commit) TypeTag {
	stats, err := commit.Stats()
	if err != nil {
		return c.DefaultType
	}

	// var filenames []string
	var extensions = make(map[string]bool)
	allTestFiles := true
	allBuildfiles := true

	for _, f := range stats {
		base := path.Base(f.Name)
		ext := path.Ext(f.Name)

		if !(strings.HasPrefix(f.Name, "test") ||
			strings.HasPrefix(base, "test")) {
			allTestFiles = false
		}
		if !(base == "Makefile" ||
			base == ".dockerignore" ||
			base == ".gitignore" ||
			strings.HasPrefix(f.Name, ".github")) {
			allBuildfiles = false
		}

		// filenames = append(filenames, f.Name)
		extensions[ext] = true
	}

	var allExtensions = make([]string, len(extensions))

	i := 0
	for k := range extensions {
		allExtensions[i] = k
		i++
	}

	log.Debug().
		Strs("extensions", allExtensions).
		Bool("all_docs", len(extensions) == 1 && extensions[".md"]).
		Bool("all_build", allBuildfiles).
		Bool("all-test", allTestFiles).
		Str("hash", commit.Hash.String()).
		Str("message", commit.Message).
		Str("message", commit.Message).
		Msg("Guessing type")

	if allTestFiles {
		return "test"
	}

	if allBuildfiles {
		return "build"
	}

	if len(extensions) == 1 && extensions[".md"] {
		return "docs"
	}

	return c.DefaultType
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
