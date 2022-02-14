package program

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"regexp"
	"strings"
)

type Changelog struct {
	Path string `arg:"" default:"."`
	Tag  string `short:"t" help:"Tag from which to start"`
}

var commitType = regexp.MustCompile("([a-zA-Z_][a-zA-Z_0-9]*):")

func (c *Changelog) Run() error {

	r, err := git.PlainOpen(c.Path)
	if err != nil {
		return err
	}

	tags, err := r.TagObjects()
	if err != nil {
		return err
	}

	defer tags.Close()

	var stopAt plumbing.Hash

	if c.Tag != "" {
		_ = tags.ForEach(func(t *object.Tag) error {
			if t.Name == c.Tag {
				stopAt = t.Hash
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

	commits := make(map[string][]*object.Commit)

	_ = iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == stopAt {
			return storer.ErrStop
		}

		re := commitType.FindStringSubmatch(commit.Message)

		tt := "fix"
		if len(re) > 1 {
			tt = re[1]
		}

		commits[tt] = append(commits[tt], commit)
		return nil
	})

	for key, list := range commits {
		fmt.Printf("%s:\n", key)
		for _, commit := range list {
			fmt.Printf("   * (%s) %s\n", commit.ID(), strings.ReplaceAll(commit.Message, "\n", "\n     "))
		}
		fmt.Println()
	}

	return nil
}
