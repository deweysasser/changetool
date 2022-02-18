package program

import (
	"bufio"
	"github.com/deweysasser/changetool/test_framework"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func SkipTestChangeLog(t *testing.T) {
	dir := test_framework.TestDir(t)
	output := path.Join(dir, "output.txt")
	r, err := test_framework.NewFromTest(t)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = r.RunFile("../changes/changeset_test_Basic.yaml")
	if err != nil {
		t.Fatal(err)
	}

	fp, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}

	opts := Options{
		Changelog: Changelog{},
		Path:      r.Path,
		OutFP:     fp,
	}

	opts.Changelog.Run(&opts)

	fp.Close()

	_, err = os.Stat(output)

	if err != nil {
		t.Fatal(err)
	}

	fp, err = os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	lines := 0
	for scanner.Scan() {
		lines++
	}

	fp.Close()

	expected := `Feature:
   * initial commit

Fix:
   * non-conventional commit comment

Docs:
   * another non-conventional commit, this time of doc

Chore:
   * do nothing real`
	bytes, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 20, lines)
	assert.Equal(t, expected, string(bytes))
}
