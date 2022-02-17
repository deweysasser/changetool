package program

import (
	"github.com/Masterminds/semver"
	"github.com/deweysasser/changetool/changes"
	"github.com/stretchr/testify/assert"
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
