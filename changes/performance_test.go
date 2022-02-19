package changes

import (
	"github.com/deweysasser/changetool/repo"
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog"
	"testing"
)

var repoPath = "../test-output/performance-test/rails"
var repoURL = "https://github.com/rails/rails.git"
var createdTagCount = 5000

func BenchmarkLargeRepos(b *testing.B) {
	gitRepo := test_framework.CloneRepo(b, repoPath, repoURL, createdTagCount)
	r, nil := repo.FromRepository(gitRepo, nil)

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(r, StopAtCount(1000), DefaultGuess("guess"))
		if err != nil {
			b.Fatal(err)
		}
	}
}
