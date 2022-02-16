package changes

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"testing"
)
import "github.com/deweysasser/changetool/test_framework"

func Test_Basic(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	r.RunFile("changeset_test_Basic.yaml")

	t.Run("full changelog", func(t *testing.T) {

		cs, err := Load(r.Repository, plumbing.ZeroHash, DefaultGuess(TypeTag("guess")))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 3, len(cs.Commits))

		writeYaml(t, cs, "changeset.yaml")
	})

	t.Run("Since Last Changelog", func(t *testing.T) {

		ref, err := r.Repository.Reference(plumbing.ReferenceName("refs/tags/v0.1"), true)
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

		cs, err := Load(r.Repository, ref.Hash(), DefaultGuess(TypeTag("guess")))

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.Equal(t, 2, len(cs.Commits))

		writeYaml(t, cs, "changeset.yaml")

	})
}

func writeYaml(t *testing.T, cs *ChangeSet, s string) {
	dir := test_framework.TestDir(t)

	bytes, err := yaml.Marshal(cs)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = os.WriteFile(path.Join(dir, s), bytes, os.ModePerm)

	if err != nil {
		assert.NoError(t, err)
		return
	}
}
