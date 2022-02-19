package program

import (
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"strings"
	"testing"
)

func TestChangeLog(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	must(t, err)

	must(t, r.RunFile("../changes/changeset_test_Basic.yaml"))

	t.Run("Basic",
		testChangelog(r.Path,
			"",
			`Feature:
   * initial commit

Fix:
   * non-conventional commit comment

Docs:
   * another non-conventional commit, this time of doc

Chore:
   * do nothing real

`))

	t.Run("Restricted tag",
		testChangelog(r.Path,
			"--since-tag v0.1",
			`Docs:
   * another non-conventional commit, this time of doc

Chore:
   * do nothing real

`))

}

func testChangelog(repo, additionalArgs, expected string) func(t *testing.T) {
	return func(t *testing.T) {
		opts := Options{}
		dir := test_framework.TestDir(t)
		output := path.Join(dir, "output.txt")

		args := []string{
			"changelog",
			"--path",
			repo,
			"--output",
			output,
		}

		if additionalArgs != "" {
			args = append(args, strings.Split(additionalArgs, " ")...)
		}

		log.Debug().Strs("args", args).Msg("Parsing")

		context, err := opts.Parse(args)

		must(t, err)

		must(t, context.Run(&opts))

		_, err = os.Stat(output)

		if err != nil {
			t.Fatal(err)
		}

		fp, err := os.Open(output)
		must(t, err)

		defer fp.Close()

		bytes, err := os.ReadFile(output)
		must(t, err)

		assert.Equal(t, expected, string(bytes))
	}
}

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
