package program

import (
	"sort"
	"strings"
)

var TypesInOrder = []string{
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
	Tag      string
	Order    int
	Messages []string
}

func asCommitList(order []string, m map[string][]string) []CommitEntry {
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

func makeEntry(order []string, k string, v []string) (entry CommitEntry) {
	entry = CommitEntry{
		Name:     strings.Title(k),
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
