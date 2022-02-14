package program

import (
	"github.com/Masterminds/semver"
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
