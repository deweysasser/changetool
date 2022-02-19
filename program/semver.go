package program

import (
	"bufio"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/deweysasser/changetool/changes"
	"github.com/deweysasser/changetool/perf"
	"github.com/deweysasser/changetool/repo"
	"github.com/deweysasser/changetool/versions"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"os"
)

type Semver struct {
	Changelog
	FromFile       string   `group:"source" xor:"source" required:"" type:"existingfile" help:"Set previous revision from the first semver looking string found in this file"`
	ReplaceIn      []string `type:"existingfile" placeholder:"FILE" help:"Replace version in these files"`
	AllowUntracked bool     `help:"allow untracked files to count as clean"`
}

func (s *Semver) Run(program *Options) error {
	r, err := program.Repository()
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

	if s.SinceTag == "" {
		s.SinceTag = foundTag
	}

	changeSet, err := s.CalculateChanges(r)
	if err != nil {
		return err
	}

	nextVersion, err2 := s.findNextVersion(version, r, changeSet)
	if err2 != nil {
		return err2
	}

	_, _ = fmt.Fprintln(program.OutFP, nextVersion.String())

	for _, f := range s.ReplaceIn {
		if err = s.ReplaceInFile(f, version.String()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Semver) findNextVersion(version semver.Version, r *repo.Repository, changes *changes.ChangeSet) (semver.Version, error) {

	status, head, err := s.gitWorktreeStatus(r)
	if err != nil {
		return semver.Version{}, err
	}

	nextVersion := version
	nextVersion, _ = nextVersion.SetPrerelease("")
	nextVersion, _ = nextVersion.SetMetadata("")

	log.Debug().
		Str("base_version", version.String()).
		Msg("Base version")

	var isClean bool

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
	} else {
		// FIXME:  this is a bit too aggressive
		isClean = status.IsClean()
	}

	log.Debug().Str("status", status.String()).Msg("working directory clean status")

	nextVersion = nextVersionFromChangeSet(changes, nextVersion)

	if !isClean {
		nextVersion = nextVersion.IncMinor()
		return nextVersion.SetPrerelease(fmt.Sprintf("dirty.%s", head.Hash().String()[:6]))
	} else {
		return nextVersion, nil
	}
}

func (s *Semver) gitWorktreeStatus(r *repo.Repository) (git.Status, *plumbing.Reference, error) {
	defer perf.Timer("getting worktree status").Stop()
	w, err := r.Worktree()

	if err != nil {
		return nil, nil, err
	}

	log.Debug().Msg("Getting status")
	status, err := w.Status()
	if err != nil {
		return nil, nil, err
	}

	log.Debug().Msg("Getting head revision")
	head, err := r.Head()
	if err != nil {
		return nil, nil, err
	}
	return status, head, nil
}

func nextVersionFromChangeSet(changes *changes.ChangeSet, version semver.Version) semver.Version {
	switch {
	case len(changes.BreakingChanges) > 0:
		log.Debug().Msg("We have breaking changes")
		// We only increment major if we're post 1.0.  Before that all changes are a "minor" level
		if version.Major() > 0 {
			version = version.IncMajor()
		} else {
			log.Debug().Msg("But we're before 1.0")
			version = version.IncMinor()
		}
	case len(changes.Commits["feat"]) > 0:
		version = version.IncMinor()
	case len(changes.Commits["fix"]) > 0:
		version = version.IncPatch()
	}

	return version
}

func (s *Semver) FindPreviousVersion(r *repo.Repository) (semver.Version, string, error) {
	if s.FromFile != "" {
		return versions.FindPreviousVersionFromFile(s.FromFile)
	} else {
		return versions.FindPreviousVersionFromTag(r)
	}
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
		line := versions.SemverRegexp.ReplaceAllString(scanner.Text(), new)
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
