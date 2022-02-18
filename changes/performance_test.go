package changes

import (
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog"
	"testing"
)

var repoPath = "../test-output/performance-test/rails"
var repoURL = "https://github.com/rails/rails.git"
var createdTagCount = 5000

func BenchmarkLargeRepos(b *testing.B) {
	repo := test_framework.CloneRepo(b, repoPath, repoURL, createdTagCount)

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(repo, StopAtCount(1000), DefaultGuess("guess"))
		if err != nil {
			b.Fatal(err)
		}
	}
}
