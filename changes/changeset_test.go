package changes

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"testing"
)
import "github.com/deweysasser/changetool/test_framework"

func Test_Basic(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = r.RunFile("changeset_test_Basic.yaml")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("default changelog", func(t *testing.T) {

		cs, err := Load(r.Repository, StopAtTagMatch(r.Repository, regexp.MustCompile(`v[0-9\.]+`).MatchString), DefaultGuess("guess"))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 1, len(cs.Commits))
		assert.Equal(t, 0, len(cs.BreakingChanges))

		writeYaml(t, cs)
	})

	t.Run("full changelog", func(t *testing.T) {

		cs, err := Load(r.Repository, StopAtCount(1000), DefaultGuess("guess"))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 3, len(cs.Commits))
		assert.Equal(t, 0, len(cs.BreakingChanges))

		writeYaml(t, cs)
	})

	t.Run("Since Last Changelog", func(t *testing.T) {

		ref, err := r.Repository.Reference("refs/tags/v0.1", true)
		if err != nil {
			assert.FailNow(t, "Failed to find assumed tag v0.1")
		}

		log.Debug().
			Str("ref", ref.String()).
			Str("name", ref.Name().String()).
			Str("hash", ref.Hash().String()[:6]).
			Msg("stopping at v0.1")

		if ref.Hash() == plumbing.ZeroHash {
			assert.FailNow(t, "Failed to find tag v0.1")
		}

		cs, err := Load(r.Repository, StopAtHash(ref.Hash()), DefaultGuess("guess"))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 2, len(cs.Commits))
		assert.Equal(t, 0, len(cs.BreakingChanges))

		writeYaml(t, cs)
	})
}

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Breaking(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	must(t, r.RunFile("changeset_test_Basic.yaml"))
	// Add in a breaking change
	must(t, r.Run([]test_framework.GitOperation{
		{
			Message: "feat!: something that breaks",
		},
	}))

	t.Run("full changelog", func(t *testing.T) {

		cs, err := Load(r.Repository, StopAtCount(1000), DefaultGuess("guess"))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 3, len(cs.Commits))
		assert.Equal(t, 1, len(cs.BreakingChanges))
		assert.Equal(t, 2, len(cs.Commits["guess"]))

		writeYaml(t, cs)
	})
}

func Test_Guessing(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	must(t, r.RunFile("changeset_test_Basic.yaml"))

	t.Run("full changelog guess", func(t *testing.T) {

		guess := func(commit *object.Commit) TypeTag {
			if t, e := StandardGuess(commit); e != nil {
				return "guess"
			} else {
				return t
			}
		}
		cs, err := Load(r.Repository, StopAtCount(1000), guess)

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 4, len(cs.Commits))
		assert.Equalf(t, 1, len(cs.Commits["docs"]), "Doc was guessed correctly")
		assert.Equalf(t, 1, len(cs.Commits["guess"]), "Should be only 1 guess")

		writeYaml(t, cs)
	})
}

func writeYaml(t *testing.T, cs *ChangeSet) {
	if !t.Failed() {
		return
	}

	bytes, err := yaml.Marshal(cs)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	_, _ = os.Stdout.Write(bytes)
}
