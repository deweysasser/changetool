package versions

import (
	"bufio"
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

var SemverRegexp = regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+)?`)

func FindPreviousVersionFromTag(r *git.Repository) (version semver.Version, foundTag string, errReturn error) {
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

func FindPreviousVersionFromFile(filename string) (semver.Version, string, error) {
	version := semver.Version{}
	// #nosec G304
	f, err := os.Open(filename)
	if err != nil {
		return version, "", err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if s := SemverRegexp.FindString(scanner.Text()); s != "" {
			if ver, err := semver.NewVersion(s); err == nil {
				return *ver, "", nil
			}
		}
	}

	return version, "", nil
}
