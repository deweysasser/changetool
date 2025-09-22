package program

import (
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog"
	"testing"
)

func BenchmarkChangelog_Run(b *testing.B) {

	var repoPath = "../test-output/program-performance/git-cli"
	var repoURL = "https://github.com/cli/cli"
	var createdTagCount = 0

	test_framework.CloneRepo(b, repoPath, repoURL, createdTagCount)

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		opts := Options{
			Changelog: Changelog{},
			Path:      repoPath,
		}

		err := opts.Changelog.Run(&opts)

		if err != nil {
			b.Fatal(err)
		}
	}
}
