package repo

import (
	"github.com/deweysasser/changetool/test_framework"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTags(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	must(t, r.RunFile("release-repo.yaml"))

	repo, _ := FromRepository(r.Repository, nil)

	tags := repo.TagMap()
	reverse := repo.ReverseTagMap()

	assert.Equal(t, 2, len(tags))

	simple := tags["v1.1"]

	c, err := r.CommitObject(simple)
	assert.NoError(t, err)
	if c != nil {
		assert.Equal(t, "non-conventional commit comment", c.Message)
		assert.Equal(t, "v1.1", reverse[c.Hash][0])
	}

	object := tags["v1.2"]

	c, err = r.CommitObject(object)
	assert.NoError(t, err)
	if c != nil {
		assert.Equal(t, "chore: do nothing real", c.Message)
		assert.Equal(t, "v1.2", reverse[c.Hash][0])
	}
}

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
