package changes

import (
	"github.com/Masterminds/semver"
	"github.com/deweysasser/changetool/repo"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
)

// StopAt is a commit recognizer
type StopAt func(commit *object.Commit) bool

// NeverStop is an StopAt that accepts nothing, ever
func NeverStop(_ *object.Commit) bool { return false }

// AlwaysStop is an StopAt that accepts every commit
func AlwaysStop(_ *object.Commit) bool { return true }

// StopAtHash stops when the hash is encountered
func StopAtHash(hash plumbing.Hash) StopAt {
	return func(commit *object.Commit) bool {
		return hash == commit.Hash
	}
}

// StopAtCount accepts the specific count of commits given
func StopAtCount(count int) StopAt {
	currentCommit := 0
	return func(commit *object.Commit) bool {
		if currentCommit > count {
			return true
		} else {
			currentCommit++
			return false
		}
	}
}

// StopAtTagMatch stops when the tag matches a tag referring to this commit
func StopAtTagMatch(r *repo.Repository, matchString func(s string) bool) StopAt {
	return func(commit *object.Commit) bool {
		for _, tag := range r.ReverseTagMap()[commit.Hash] {
			if matchString(tag) {
				return true
			}
		}
		return false
	}
}

func StopAtFirstSemver(r *repo.Repository) StopAt {
	return func(commit *object.Commit) bool {
		for _, tag := range r.ReverseTagMap()[commit.Hash] {
			log.Debug().Str("tag", tag).Msg("checking tag")
			if _, err := semver.NewVersion(tag); err == nil {
				log.Debug().
					Str("hash", commit.Hash.String()[:6]).
					Str("tag", tag).
					Msg("Stopping at tagged commit")
				return true
			}
		}
		return false
	}
}
