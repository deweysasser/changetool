package versions

import (
	"bufio"
	"github.com/Masterminds/semver"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

var SemverRegexp = regexp.MustCompile(`v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+)?`)

func FindPreviousVersionFromTag(r *repo.Repository) (version semver.Version, foundTag string, errReturn error) {
	version = semver.Version{}
	log.Debug().Msg("finding previous version by examining tags")

	commits, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return version, "", err
	}
	defer commits.Close()

	log.Debug().Msg("Examining commits")

	reverseTagMap := r.ReverseTagMap()

	_ = commits.ForEach(func(commit *object.Commit) error {

		log.Debug().
			Str("hash", commit.Hash.String()[:6]).
			Msg("Looking up commit")
		for _, tag := range reverseTagMap[commit.Hash] {
			log.Debug().Str("tag", tag).Str("regex", SemverRegexp.String()).Msg("Matching against tag")
			if v, err := semver.NewVersion(tag); err == nil {
				if v.GreaterThan(&version) {
					version = *v
					foundTag = tag
				}
			}
		}

		if foundTag != "" {
			return storer.ErrStop
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
