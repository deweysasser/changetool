package program

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type Changelog struct {
	Path        string   `arg:"" default:"."`
	Tag         string   `short:"t" help:"Tag from which to start"`
	DefaultType string   `default:"fix" help: if type is not specified in commit, assume this type`
	Order       []string `default:"${type_order}" help: order in which to list commit message types`
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

	var stopAt plumbing.Hash

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

				stopAt = t.Hash()
				return storer.ErrStop
			}

			return nil
		})
	}
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

		tt := c.DefaultType

		if len(re) > 1 {
			tt = re[1]
			message = message[len(re[0]):]
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
