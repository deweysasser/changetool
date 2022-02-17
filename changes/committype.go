package changes

import (
	"errors"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"path"
	"sort"
	"strings"
)

// TypeTag represents the type of a commit.  There is a pre-defined set, but it can also be dynamically extended.
type TypeTag string

// Types is a list of TypeTag
type Types []TypeTag

// TypesInOrder is the standard ordering of types
var TypesInOrder = Types{
	"feat",
	"fix",
	"test",
	"docs",
	"build",
	"refactor",
	"chore",
}

var NoClue TypeTag = "--no clue--"

// CommitTypeEntry represents a list of messages of a specific type
type CommitTypeEntry struct {
	Name     string
	Tag      TypeTag
	Order    int
	Messages []string
}

// Join joins the specified Types using the given separator
func (t Types) Join(sep string) string {
	strs := make([]string, len(t))

	for n, s := range t {
		strs[n] = string(s)
	}

	return strings.Join(strs, sep)
}

// CommitEntries returns a list of CommitTypeEntry
func CommitEntries(order []TypeTag, m map[TypeTag][]string) []CommitTypeEntry {
	var list []CommitTypeEntry

	for k, v := range m {
		list = append(list, makeEntry(order, k, v))
	}

	inOrder := func(i, j int) bool {
		switch {
		case list[i].Order < list[j].Order:
			return true
		case list[i].Order > list[j].Order:
			return false
		default:
			return list[i].Name < list[j].Name
		}
	}

	sort.Slice(list, inOrder)

	return list
}

func makeEntry(order []TypeTag, k TypeTag, v []string) (entry CommitTypeEntry) {
	entry = CommitTypeEntry{
		Name:     strings.Title(string(k)),
		Tag:      k,
		Order:    1000,
		Messages: v,
	}

	if entry.Tag == "feat" {
		entry.Name = "Feature"
	}

	for n, t := range order {
		if t == k {
			entry.Order = n
			return
		}
	}

	return
}

// StandardGuess guesses the commit type by examining the commit
func StandardGuess(commit *object.Commit) (TypeTag, error) {
	stats, err := commit.Stats()
	if err != nil {
		return NoClue, err
	}

	// var filenames []string
	var extensions = make(map[string]bool)
	allTestFiles := true
	allBuildfiles := true

	for _, f := range stats {
		base := path.Base(f.Name)
		ext := path.Ext(f.Name)

		if !(strings.HasPrefix(f.Name, "test") ||
			strings.HasPrefix(base, "test")) {
			allTestFiles = false
		}
		if !(base == "Makefile" ||
			base == ".dockerignore" ||
			base == ".gitignore" ||
			strings.HasPrefix(f.Name, ".github")) {
			allBuildfiles = false
		}

		// filenames = append(filenames, f.Name)
		extensions[ext] = true
	}

	var allExtensions = make([]string, len(extensions))

	i := 0
	for k := range extensions {
		allExtensions[i] = k
		i++
	}

	log.Debug().
		Strs("extensions", allExtensions).
		Bool("all_docs", len(extensions) == 1 && extensions[".md"]).
		Bool("all_build", allBuildfiles).
		Bool("all-test", allTestFiles).
		Str("hash", commit.Hash.String()).
		Str("message", commit.Message).
		Str("message", commit.Message).
		Msg("Guessing type")

	if allTestFiles {
		return "test", nil
	}

	if allBuildfiles {
		return "build", nil
	}

	if len(extensions) == 1 && extensions[".md"] {
		return "docs", nil
	}
	return NoClue, errors.New("no guess")
}
