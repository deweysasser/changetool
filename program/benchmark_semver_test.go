package program

import (
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func BenchmarkSemver_Run(b *testing.B) {

	var repoPath = "../test-output/program-performance/git-cli"
	var repoURL = "https://github.com/cli/cli"
	var createdTagCount = 0

	test_framework.CloneRepo(b, repoPath, repoURL, createdTagCount)

	dir := test_framework.TestDir(b)
	output := path.Join(dir, "version.txt")

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		opts := Options{
			Semver: Semver{
				AllowUntracked: true,
			},
			Path:   repoPath,
			Output: output,
		}

		opts.AfterApply()
		err := opts.Semver.Run(&opts)

		if err != nil {
			b.Fatal(err)
		}

		bytes, err := os.ReadFile(output)
		if err != nil {
			b.Fatal(err)
		}

		assert.Equal(b, "2.6", string(bytes))
	}
}
