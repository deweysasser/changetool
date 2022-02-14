package program

import (
	"bufio"
	"fmt"
	"github.com/Masterminds/semver"
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
	FromTag        bool     `group:"source" xor:"source" required:"" help:"Set semver from the last tag" `
	FromFile       string   `group:"source" xor:"source" required:"" type:"existingfile" help:"Set previous revision from the first semver looking string found in this file"`
	ReplaceIn      []string `type:"existingfile" placeholder:"FILE" help:"Replace version in these files"`
	AllowUntracked bool     `help:"allow untracked files to count as clean"`
}

func (s *Semver) Run() error {
	r, err := git.PlainOpen(s.Path)
	if err != nil {
		return err
	}
	version, foundTag, err := s.FindPreviousVersion(r)

	if err != nil {
		return err
	}

	log.Debug().
		Str("previous_version", version.String()).
		Msg("Found previous version")

	if s.FromTag && s.SinceTag == "" {
		s.SinceTag = foundTag
	}

	changes, err := s.CalculateChanges()
	if err != nil {
		return err
	}

	nextVersion := version
	log.Debug().
		Str("base_version", version.String()).
		Msg("Base version")

	nextVersion, _ = nextVersion.SetPrerelease("")
	nextVersion, _ = nextVersion.SetMetadata("")

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

	// FIXME:  this is a bit too aggressive
	isClean := status.IsClean()

	if s.AllowUntracked {
		clean := true

		for f, file := range status {
			if file.Worktree != git.Untracked {
				clean = false
			}
			log.Debug().Str("file", f).
				Int8("worktree_status", int8(file.Worktree)).
				Int8("staging_status", int8(file.Staging)).
				Msg("File status")
		}

		isClean = clean
	}
	switch {
	case len(changes.BreakingChanges) > 0:
		log.Debug().Msg("We have breaking changes")
		// We only increment major if we're post 1.0.  Before that all changes are a "minor" level
		if version.Major() > 0 {
			nextVersion = nextVersion.IncMajor()
		} else {
			log.Debug().Msg("But we're before 1.0")
			nextVersion = nextVersion.IncMinor()
		}
	case !isClean:
		log.Debug().Str("status", status.String()).Msg("working directory not clean")
		nextVersion = nextVersion.IncMinor()
	case len(changes.Commits["feat"]) > 0:
		nextVersion = nextVersion.IncMinor()
	case len(changes.Commits["fix"]) > 0:
		nextVersion = nextVersion.IncPatch()
	}

	// We want to append this in any case when the worktree is dirty
	if !isClean {
		if nextVersion, err = nextVersion.SetPrerelease(fmt.Sprintf("dirty.%s", head.Hash().String()[:6])); err != nil {
			return err
		}
	}

	fmt.Println(nextVersion.String())

	for _, f := range s.ReplaceIn {
		if err = s.ReplaceInFile(f, version.String()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Semver) FindPreviousVersion(r *git.Repository) (semver.Version, string, error) {
	if s.FromTag {
		return s.FindPreviousVersionFromTag(r)
	} else {
		return s.FindPreviousVersionFromFile()
	}
}

func (s *Semver) FindPreviousVersionFromTag(r *git.Repository) (version semver.Version, foundTag string, errReturn error) {
	version = semver.Version{}
	allTags := make(map[plumbing.Hash]plumbing.ReferenceName)
	log.Debug().Msg("finding previous version by examining tags")

	tags, err := r.Tags()
	if err != nil {
		return version, foundTag, err
	}

	defer tags.Close()

	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		log.Debug().
			Str("name", ref.Name().String()).
			Str("ref", ref.Hash().String()[:6]).
			Msg("Examining simple tag")

		allTags[ref.Hash()] = ref.Name()
		return nil
	})

	tagObjects, err := r.TagObjects()
	if err != nil {
		return version, foundTag, err
	}

	defer tagObjects.Close()

	_ = tagObjects.ForEach(func(ref *object.Tag) error {
		// Should this be converted toa reference name, or should we change the map type?
		if commit, err := ref.Commit(); err != nil {
			log.Err(err).Str("name", ref.Name).Msg("Error examining commit for hash")
			return storer.ErrStop
		} else {
			log.Debug().
				Str("name", ref.Name).
				Str("ref", commit.Hash.String()[:6]).
				Msg("Examining tag object")
			allTags[commit.Hash] = plumbing.ReferenceName(ref.Name)
			return nil
		}
	})

	commits, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return version, "", err
	}
	defer commits.Close()

	log.Debug().Msg("Examining commits")

	_ = commits.ForEach(func(commit *object.Commit) error {
		log.Debug().
			Str("hash", commit.Hash.String()[:6]).
			Msg("Checking commit hash")
		if name, exists := allTags[commit.Hash]; exists {
			log.Debug().
				Str("tag", name.Short()).
				Msg("found matching tag")
			parseName := strings.TrimPrefix(name.Short(), "v")
			// Should we look for anything with the right regexp?  Or stick to the "v*" convention?

			log.Debug().Str("name", parseName).Msg("parsing")
			if v, err := semver.NewVersion(parseName); err == nil {
				version = *v
				foundTag = name.Short()
				return storer.ErrStop
			} else {
				log.Debug().AnErr("error", err).Msg("Error parsing tag for semver")
			}
		}

		return nil
	})

	return version, foundTag, nil
}

var semverRegexp = regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+)?`)

func (s *Semver) FindPreviousVersionFromFile() (semver.Version, string, error) {
	version := semver.Version{}
	f, err := os.Open(s.FromFile)
	if err != nil {
		return version, "", err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if s := semverRegexp.FindString(scanner.Text()); s != "" {
			if ver, err := semver.NewVersion(s); err == nil {
				return *ver, "", nil
			}
		}
	}

	return version, "", nil
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
