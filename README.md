# concom

A tool for working with projects that use the [conventional commit](https://www.conventionalcommits.org/) conventions.

## Overview

This is a first, very simple, pass at a tool for working with conventional commits projects.

The other tools that I've found are unsatisfactory in some way, either by having *way* too many
dependencies (I'm looking gat you, Node based tools) or by producing output that I don't really want,
or by requiring templates and config in the project.

## Changelogs

```text
Usage: concom changelog [<path>]

Arguments:
  [<path>]

Flags:
  -h, --help                              Show context-sensitive help.
  -d, --debug                             Show debugging information
  -l, --log-format="auto"                 How to show program output (auto|terminal|jsonl)
  -q, --quiet                             Be less verbose than usual

  -t, --tag=STRING                        Tag from which to start
      --default-type="fix"                if type is not specified in commit, assume this type
      --[no-]guess-missing-commit-type    If commit type is missing, take a guess about which it is
      --order=feat,fix,test,docs,build,refactor,chore,...
                                          order in which to list commit message types
```

## Version bumps

to be implemented