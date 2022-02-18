package changes

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"testing"
)

var repoPath = "../test-output/performance-test/rails"
var repoURL = "https://github.com/rails/rails.git"
var createdTagCount = 5000

func BenchmarkLargeRepos(b *testing.B) {

	var repo *git.Repository

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		_ = os.MkdirAll(path.Dir(repoPath), os.ModePerm|os.ModeDir)
		log.Debug().Str("url", repoURL).Msg("Cloning Repo")
		repo, err = git.PlainClone(repoPath, true, &git.CloneOptions{URL: repoURL})
		if err != nil {
			b.Fatal(err)
		}

		// Rails doesn't have as many tags as we want, so let's create a few thousand
		var commits []plumbing.Hash
		if iter, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime}); err != nil {
			b.Fatal(err)
		} else {
			iter.ForEach(func(commit *object.Commit) error {
				commits = append(commits, commit.Hash)
				return nil
			})
		}

		log.Debug().Int("count", createdTagCount).Msg("Creating extra tags")
		tagEvery := len(commits) / createdTagCount
		for i := 0; i < len(commits); i += tagEvery {
			repo.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName(fmt.Sprintf("refs/tags/scale-test-tag-%d", i)), commits[i]))
		}

	} else {
		log.Debug().Msg("Using existing repo")

		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			b.Fatal(err)
		}
	}

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(repo, plumbing.ZeroHash, DefaultGuess("guess"))
		if err != nil {
			b.Fatal(err)
		}
	}
}
