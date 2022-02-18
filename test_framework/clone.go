package test_framework

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"testing"
	"time"
)

func CloneRepo(b *testing.B, repoPath, repoURL string, createdTagCount int) *git.Repository {

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		_ = os.MkdirAll(path.Dir(repoPath), os.ModePerm|os.ModeDir)
		log.Debug().Str("url", repoURL).Msg("Cloning Repo")
		now := time.Now()

		repo, err := git.PlainClone(repoPath, true, &git.CloneOptions{URL: repoURL})
		if err != nil {
			b.Fatal(err)
		}

		log.Debug().Dur("clone_time", time.Since(now)).Msg("Repo cloned")

		// Rails doesn't have as many tags as we want, so let's create a few thousand
		log.Debug().Msg("Creating tags")
		now = time.Now()
		var commits []plumbing.Hash
		if iter, err := repo.Log(&git.LogOptions{}); err != nil {
			b.Fatal(err)
		} else {
			if err := iter.ForEach(func(commit *object.Commit) error {
				commits = append(commits, commit.Hash)
				return nil
			}); err != nil {
				b.Fatal(err)
			}
		}

		if createdTagCount > 0 {
			log.Debug().Int("count", createdTagCount).Msg("Creating extra tags")
			tagEvery := len(commits) / createdTagCount
			for i := 0; i < len(commits); i += tagEvery {
				if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName(fmt.Sprintf("refs/tags/scale-test-tag-%d", i)), commits[i])); err != nil {
					b.Fatal(err)
				}
			}

			log.Debug().Dur("tag_time", time.Since(now)).Msg("Tags Created")
		}
		return repo
	} else {
		log.Debug().Msg("Using existing repo")

		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			b.Fatal(err)
		}

		return repo
	}
}
