package versions

import (
	"github.com/deweysasser/changetool/repo"
	"github.com/deweysasser/changetool/test_framework"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Prerelease(t *testing.T) {
	r1, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.NoError(t, r1.RunFile("prerelease-repo.yaml"))

	r, _ := repo.FromRepository(r1.Repository, nil)

	t.Run("full changelog", func(t *testing.T) {

		ver, tag, err := FindPreviousVersionFromTag(r)
		assert.NoError(t, err)
		assert.Equal(t, "v0.2", tag)
		assert.Equal(t, int64(0), ver.Major())
		assert.Equal(t, int64(2), ver.Minor())
		assert.Equal(t, int64(0), ver.Patch())
		assert.Equal(t, "", ver.Prerelease())
		assert.Equal(t, "", ver.Metadata())
	})

	t.Run("With breaking change", func(t *testing.T) {

		must(t, r1.Run([]test_framework.GitOperation{
			{Message: "feat!: something that breaks"},
		}))

		ver, tag, err := FindPreviousVersionFromTag(r)
		assert.NoError(t, err)
		assert.Equal(t, "v0.2", tag)
		assert.Equal(t, int64(0), ver.Major())
		assert.Equal(t, int64(2), ver.Minor())
		assert.Equal(t, int64(0), ver.Patch())
		assert.Equal(t, "", ver.Prerelease())
		assert.Equal(t, "", ver.Metadata())
	})

}
func Test_Post_1_0(t *testing.T) {
	r1, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.NoError(t, r1.RunFile("release-repo.yaml"))

	r, _ := repo.FromRepository(r1.Repository, nil)

	t.Run("full changelog", func(t *testing.T) {

		ver, tag, err := FindPreviousVersionFromTag(r)
		assert.NoError(t, err)
		assert.Equal(t, "v1.2", tag)
		assert.Equal(t, int64(1), ver.Major())
		assert.Equal(t, int64(2), ver.Minor())
		assert.Equal(t, int64(0), ver.Patch())
		assert.Equal(t, "", ver.Prerelease())
		assert.Equal(t, "", ver.Metadata())
	})

	t.Run("With breaking change", func(t *testing.T) {

		must(t, r1.Run([]test_framework.GitOperation{
			{Message: "feat!: something that breaks"},
		}))

		ver, tag, err := FindPreviousVersionFromTag(r)
		assert.NoError(t, err)
		assert.Equal(t, "v1.2", tag)
		assert.Equal(t, int64(1), ver.Major())
		assert.Equal(t, int64(2), ver.Minor())
		assert.Equal(t, int64(0), ver.Patch())
		assert.Equal(t, "", ver.Prerelease())
		assert.Equal(t, "", ver.Metadata())
	})

}

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
