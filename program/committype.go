package program

import (
	"sort"
	"strings"
)

type TypeTag string

type Types []TypeTag

var TypesInOrder = Types{
	"feat",
	"fix",
	"test",
	"docs",
	"build",
	"refactor",
	"chore",
}

type CommitEntry struct {
	Name     string
	Tag      TypeTag
	Order    int
	Messages []string
}

func (t Types) Join(sep string) string {
	strs := make([]string, len(t))

	for n, s := range t {
		strs[n] = string(s)
	}

	return strings.Join(strs, sep)
}

func asCommitList(order []TypeTag, m map[TypeTag][]string) []CommitEntry {
	var list []CommitEntry

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

func makeEntry(order []TypeTag, k TypeTag, v []string) (entry CommitEntry) {
	entry = CommitEntry{
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
