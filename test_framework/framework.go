package test_framework

import (
	"errors"
	"fmt"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"strconv"
)

type GitOperation struct {
	Message string   `yaml:"message"`
	Files   []string `yaml:"files"`
	Tag     string   `yaml:"tag"`
}

type MyRepo struct {
	*repo.Repository
	Path string
}

// RunFile loads a YAML file of git operations into the repository
func (r *MyRepo) RunFile(file string) error {
	ops := make([]GitOperation, 0)

	// #nosec G304
	if bytes, err := os.ReadFile(file); err != nil {
		return err
	} else {

		err = yaml.Unmarshal(bytes, &ops)
		if err != nil {
			return err
		}
	}
	return r.Run(ops)
}

// Run runs the operations on this repo
func (r MyRepo) Run(ops []GitOperation) error {

	for n, op := range ops {

		switch {
		case op.Tag != "":
			if err := r.RunTag(op); err != nil {
				return err
			}
		case op.Message != "":
			if err := r.RunCommit(op, n); err != nil {
				return fmt.Errorf("error creating commit: %w", err)
			}
		default:
			return errors.New("don't know how to handle op " + strconv.Itoa(n))
		}
	}

	return nil
}

func (r MyRepo) RunTag(op GitOperation) error {
	// If tag has a message, create a full tag
	h, err := r.Head()
	if err != nil {
		return fmt.Errorf("error finding head: %w", err)
	}

	if op.Message != "" {

		log.Debug().Str("hash", h.Hash().String()[:6]).
			Str("name", op.Tag).
			Msg("Creating tag object")
		if _, err := r.CreateTag(op.Tag, h.Hash(), &git.CreateTagOptions{Message: op.Message}); err != nil {
			return fmt.Errorf("error creating tag object for %s: %w", op.Tag, err)
		}
	} else {
		log.Debug().Str("hash", h.Hash().String()[:6]).
			Str("name", op.Tag).
			Msg("Creating lightweight tag")
		if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/"+op.Tag), h.Hash())); err != nil {
			return fmt.Errorf("error creating lightweight tag: %w", err)
		}
	}
	return nil
}

func (r MyRepo) RunCommit(op GitOperation, n int) error {

	if len(op.Files) == 0 {
		op.Files = append(op.Files, "example-file.c")
	}

	w, err := r.Repository.Worktree()
	if err != nil {
		return err
	}

	for _, file := range op.Files {
		filePath := path.Join(r.Path, file)
		// #nosec G304
		fp, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}

		_, err = fp.WriteString(fmt.Sprintf("File content for step %d\n", n))

		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", file, err)
		}

		if err = fp.Close(); err != nil {
			return fmt.Errorf("error closing file: %w", err)
		}
		_, err = w.Add(file)
		if err != nil {
			return fmt.Errorf("error adding file %s in %s: %w", file, r.Path, err)
		}
	}

	_, err = w.Commit(op.Message, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("error making commit: %w", err)
	}

	return nil
}

type Namer interface {
	Name() string
}

// TestDir returns the appropriate test dir, creating it if necessary
func TestDir(t Namer) string {
	base := path.Join("../test-output", t.Name())
	info, err := os.Stat(base)
	switch {
	case os.IsNotExist(err):
		// #nosec G304

		if err := os.MkdirAll(base, os.ModePerm|os.ModeDir); err != nil {
			panic(err)
		} else {
			return base
		}
	case err != nil:
		panic(err) // we want to panic here to maximize convenience for using this method
	case info.IsDir():
		return base
	default:
		panic("Unknown situation")
	}
}

// NewFromTest returns a repo in an output directory, and the directory
func NewFromTest(t Namer) (*MyRepo, error) {
	return New(path.Join(TestDir(t), "repo"))
}

func New(path string) (*MyRepo, error) {

	// Wipe it out if necessary
	_ = os.RemoveAll(path)

	err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, err
	}
	if repo, err := repo.FromRepository(git.PlainInit(path, false)); err != nil {
		return nil, err
	} else {
		c, err := repo.Config()
		if err != nil {
			return nil, err
		}
		c.Author = struct{ Name, Email string }{"ChangeTool Testing", "testing@example.com"}
		err = repo.SetConfig(c)
		if err != nil {
			return nil, err
		}
		return &MyRepo{Repository: repo, Path: path}, nil
	}
}
