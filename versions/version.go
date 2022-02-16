package versions

import "github.com/Masterminds/semver"

// Version represents the semantic version number
type Version struct {
	semver.Version
	// Tag is (optionally) the tag uses for this version
	Tag string
}
