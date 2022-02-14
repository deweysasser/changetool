package program

import (
	"bufio"
	"fmt"
	"github.com/blang/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
	"strings"
)

type Semver struct {
	Changelog
	FromTag   bool     `group:"source" xor:"source" required:"" help:"Set semver from the last tag" `
	FromFile  string   `group:"source" xor:"source" required:"" type:"existingfile" help:"Set previous revision from the first semver looking string found in this file"`
	ReplaceIn []string `type:"existingfile" help:"Replace version in these files"`
}

func (s *Semver) Run() error {
	r, err := git.PlainOpen(s.Path)
	if err != nil {
		return err
	}
	version, err := s.FindPreviousVersion(r)

	if err != nil {
		return err
	}

	log.Debug().
		Str("previous_version", version.String()).
		Msg("Found previous version")

	changes, err := s.CalculateChanges()
	if err != nil {
		return err
	}

	nextVersion := version

	nextVersion.Build = []string{}
	nextVersion.Pre = []semver.PRVersion{}

	log.Debug().Msg("Getting worktree")
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	log.Debug().Msg("Getting status")
	status, err := w.Status()
	if err != nil {
		return err
	}

	log.Debug().Msg("Getting head revision")
	head, err := r.Head()
	if err != nil {
		return err
	}

	isClean := status.IsClean()

	switch {
	case len(changes.BreakingChanges) > 0:
		nextVersion.Major = version.Major + 1
		nextVersion.Minor = 0
		nextVersion.Patch = 0
	case !isClean:
		log.Debug().Str("status", status.String()).Msg("working directory not clean")
		nextVersion.Minor = version.Minor + 1
		nextVersion.Patch = 0
	case len(changes.Commits["feat"]) > 0:
		nextVersion.Minor = version.Minor + 1
		nextVersion.Patch = 0
	case len(changes.Commits["fix"]) > 0:
		nextVersion.Patch = version.Patch + 1
	}

	// We want to append this in any case when the worktree is dirty
	if !isClean {
		nextVersion.Pre = append(nextVersion.Pre, semver.PRVersion{VersionStr: fmt.Sprintf("dirty.%s", head.Hash().String()[:6])})
	}

	fmt.Println(nextVersion.String())

	for _, f := range s.ReplaceIn {
		if err = s.ReplaceInFile(f, version.String()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Semver) FindPreviousVersion(r *git.Repository) (semver.Version, error) {
	if s.FromTag {
		return s.FindPreviousVersionFromTag(r)
	} else {
		return s.FindPreviousVersionFromFile()
	}
}

func (s *Semver) FindPreviousVersionFromTag(r *git.Repository) (version semver.Version, errReturn error) {
	version = semver.Version{}
	allTags := make(map[plumbing.Hash]plumbing.ReferenceName)

	tags, err := r.Tags()
	if err != nil {
		return version, err
	}

	defer tags.Close()
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		allTags[ref.Hash()] = ref.Name()
		return nil
	})

	commits, err := r.Log(&git.LogOptions{})
	if err != nil {
		return version, err
	}
	defer commits.Close()

	_ = commits.ForEach(func(commit *object.Commit) error {
		if name, exists := allTags[commit.Hash]; exists {
			log.Debug().
				Str("tag", name.Short()).
				Msg("Checking tag")
			parseName := strings.TrimPrefix(name.Short(), "v")
			// Should we look for anything with the right regexp?  Or stick to the "v*" convention?

			if version, err = semver.Parse(parseName); err == nil {
				return storer.ErrStop
			}
		}

		return nil
	})

	return version, nil
}

var semverRegexp = regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+)?`)

func (s *Semver) FindPreviousVersionFromFile() (semver.Version, error) {
	version := semver.Version{}
	f, err := os.Open(s.FromFile)
	if err != nil {
		return version, err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if s := semverRegexp.FindString(scanner.Text()); s != "" {
			if ver, err := semver.Parse(s); err == nil {
				return ver, nil
			}
		}
	}

	return version, nil
}

func (s *Semver) ReplaceInFile(filename string, new string) error {

	// #nosec G304
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	tmp := filename + ".tmp"
	// #nosec G304
	out, err := os.Create(tmp)

	if err != nil {
		return err
	}

	defer func() {
		_ = out.Close()
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := semverRegexp.ReplaceAllString(scanner.Text(), new)
		if _, err = out.WriteString(line); err != nil {
			return err
		}
		if _, err = out.WriteString("\n"); err != nil {
			return err
		}
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = out.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, filename)
}
