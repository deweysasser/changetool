package program

import (
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
	Path                   string   `arg:"" default:"."`
	Tag                    string   `short:"t" help:"Tag from which to start"`
	DefaultType            string   `default:"fix" help:"if type is not specified in commit, assume this type"`
	GuessMissingCommitType bool     `default:"true" negatable:"" help:"If commit type is missing, take a guess about which it is"`
	Order                  []string `default:"${type_order}" help:"order in which to list commit message types"`
}

var commitType = regexp.MustCompile("([a-zA-Z_][a-zA-Z_0-9]*): *")

func (c *Changelog) Run() error {

	r, err := git.PlainOpen(c.Path)
	if err != nil {
		return err
	}

	tags, err := r.Tags()
	if err != nil {
		return err
	}

	defer tags.Close()

	stopAt := c.findStartVersion(tags)
	iter, err := r.Log(&git.LogOptions{})

	if err != nil {
		return err
	}

	defer iter.Close()

	commits := make(map[string][]string)

	_ = iter.ForEach(func(commit *object.Commit) error {
		if len(commit.ParentHashes) > 1 {
			return nil
		}

		if commit.Hash == stopAt {
			return storer.ErrStop
		}

		message := commit.Message
		re := commitType.FindStringSubmatch(message)

		var tt string

		switch {
		case len(re) > 1:
			tt = re[1]
			message = message[len(re[0]):]
		case c.GuessMissingCommitType:
			tt = c.guessType(commit)
		default:
			tt = c.DefaultType
		}

		commits[tt] = append(commits[tt], message)
		return nil
	})

	for _, section := range asCommitList(c.Order, commits) {
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

// guessType guesses the type of the commit from information in the commit
func (c *Changelog) guessType(commit *object.Commit) string {
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

func (c *Changelog) findStartVersion(tags storer.ReferenceIter) (stop plumbing.Hash) {
	if c.Tag != "" {
		lookFor := "refs/tags/" + c.Tag
		log.Debug().Str("tag", c.Tag).Msg("Looking for tag")
		_ = tags.ForEach(func(t *plumbing.Reference) error {
			name := t.Name().String()
			log.Debug().
				Str("tag", name).Msg("comparing tag")
			if name == lookFor {
				log.Debug().
					Str("tag", c.Tag).
					Str("hash", t.Hash().String()).
					Msg("Found hash for tag")

				stop = t.Hash()
				return storer.ErrStop
			}

			return nil
		})
	}

	return plumbing.Hash{}
}
