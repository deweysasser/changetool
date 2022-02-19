package program

import (
	"github.com/Masterminds/semver"
	"github.com/deweysasser/changetool/changes"
	"github.com/deweysasser/changetool/test_framework"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"strings"
	"testing"
)

func Test_SemverMastermind(t *testing.T) {

	tests := []struct {
		name                string
		in                  string
		major, minor, patch int
		prerelease          string
	}{
		{"Basic", "1.2.3", 1, 2, 3, ""},
		{"Basic short", "1.2", 1, 2, 0, ""},
		{"Basic short prerelease", "1.2-alpha.1", 1, 2, 0, "alpha.1"},
		{"Prerelease long", "0.1.2", 0, 1, 2, ""},
		{"Prerelease short", "0.1", 0, 1, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := semver.NewVersion(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.major, int(version.Major()))
			assert.Equal(t, tt.minor, int(version.Minor()))
			assert.Equal(t, tt.patch, int(version.Patch()))
			assert.Equal(t, tt.prerelease, version.Prerelease())
		})
	}
}

func makeVersion(t *testing.T, ver string) semver.Version {
	v, err := semver.NewVersion(ver)
	if err != nil {
		t.Fatal(err)
	}

	return *v
}

func Test_nextVersionFromChangeset(t *testing.T) {
	tests := []struct {
		name    string
		changes *changes.ChangeSet
		version semver.Version
		want    semver.Version
	}{
		{
			name:    "first revision",
			changes: &changes.ChangeSet{},
			version: makeVersion(t, "0.0.0"),
			want:    makeVersion(t, "0.0.0"),
		},
		{
			name:    "Only fixes",
			changes: &changes.ChangeSet{Commits: map[changes.TypeTag][]string{"fix": []string{"a fix"}}},
			version: makeVersion(t, "0.0.0"),
			want:    makeVersion(t, "0.0.1"),
		},
		{
			name:    "Only Features",
			changes: &changes.ChangeSet{Commits: map[changes.TypeTag][]string{"feat": []string{"a fix"}}},
			version: makeVersion(t, "0.0.0"),
			want:    makeVersion(t, "0.1.0"),
		},
		{
			name:    "Features on features",
			changes: &changes.ChangeSet{Commits: map[changes.TypeTag][]string{"feat": []string{"a fix"}}},
			version: makeVersion(t, "0.1.1"),
			want:    makeVersion(t, "0.2.0"),
		},
		{
			name:    "Breaking changes on pre-release",
			changes: &changes.ChangeSet{BreakingChanges: []string{"breaking"}, Commits: map[changes.TypeTag][]string{"feat": []string{"a fix"}}},
			version: makeVersion(t, "0.1.1"),
			want:    makeVersion(t, "0.2.0"),
		},
		{
			name:    "Breaking changes on release",
			changes: &changes.ChangeSet{BreakingChanges: []string{"breaking"}, Commits: map[changes.TypeTag][]string{"feat": []string{"a fix"}}},
			version: makeVersion(t, "1.1.1"),
			want:    makeVersion(t, "2.0.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, nextVersionFromChangeSet(tt.changes, tt.version), "nextVersionFromChangeSet(%v, %v)", tt.changes, tt.version)
		})
	}
}

func TestSemverPrerelease(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	must(t, err)

	must(t, r.RunFile("../changes/changeset_test_Basic.yaml"))

	t.Run("Basic",
		testSemver(r.Path,
			"--from-tag",
			"0.2.0\n"))

	must(t, r.RunCommit(test_framework.GitOperation{Message: "fix: added a fix"}, 0))

	t.Run("With fix",
		testSemver(r.Path,
			"--from-tag",
			"0.2.1\n"))

	must(t, r.RunCommit(test_framework.GitOperation{Message: "feat: added a feat"}, 0))

	t.Run("With feat",
		testSemver(r.Path,
			"--from-tag",
			"0.3.0\n"))

}

func TestSemverPostrelease(t *testing.T) {
	r, err := test_framework.NewFromTest(t)
	must(t, err)

	must(t, r.RunFile("../versions/release-repo.yaml"))

	t.Run("Basic",
		testSemver(r.Path,
			"--from-tag",
			"1.2.0\n"))

	must(t, r.RunCommit(test_framework.GitOperation{Message: "fix: added a fix"}, 0))

	t.Run("With fix",
		testSemver(r.Path,
			"--from-tag",
			"1.2.1\n"))

	must(t, r.RunCommit(test_framework.GitOperation{Message: "feat: added a feat"}, 0))

	t.Run("With feat",
		testSemver(r.Path,
			"--from-tag",
			"1.3.0\n"))

	must(t, r.RunCommit(test_framework.GitOperation{Message: "feat!: break the world"}, 0))

	t.Run("With feat",
		testSemver(r.Path,
			"--from-tag",
			"2.0.0\n"))

}

func testSemver(repo, additionalArgs, expected string) func(t *testing.T) {
	return func(t *testing.T) {
		opts := Options{}
		dir := test_framework.TestDir(t)
		output := path.Join(dir, "output.txt")

		args := []string{
			"semver",
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
