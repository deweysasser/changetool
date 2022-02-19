# changetool

A tool for working with projects that use the [conventional commits](https://www.conventionalcommits.org/).

Generates changelogs and calculates semantic versions from tags and commit messages.

## Quickstart

Create a changelog since the last version tag: 
```shell
changetool changelog
```

Create a changelog since a specific version tag: 
```shell
changetool changelog --since-tag v1.0
```

Find the next semantic version, based on changelog: 
```shell
changetool semver
```

Update a file with the version: 
```shell
changetool semver --replace-in version.go
```

## Status

Becoming useful.

This tool should be able to generate useful changelogs/release notes and do intelligent semantic version calculations.

It still needs some work to be a low friction, very useful tool.  Suggestions on how to make it such are welcome.

## Overview

This is a first, very simple, pass at a tool for working with conventional commits projects.

The other tools that I've found are unsatisfactory in some way, either by having *way* too many
dependencies (I'm looking at you, Node based tools) or by producing output that I don't really want,
or by requiring templates and config in the project.

## Changelog generation

```text
Usage: changetool changelog

calculate changelogs

Flags:
  -h, --help    Show context-sensitive help.

info
  -d, --debug                Show debugging information
  -l, --log-format="auto"    How to show program output (auto|terminal|jsonl)
  -q, --quiet                Be less verbose than usual

locations
  -p, --path="."      Path for the git worktree/repo to log
  -o, --output="-"    File to which to send output

source
  -n, --max-commits=1000    max number of commits to check
  -s, --since-tag=STRING    Tag from which to start
  -a, --all-commits         report changelog on all commits up to --max-commits. Otherwise, report only to last version tag

calculation
  --default-type="fix"                if type is not specified in commit, assume this type
  --[no-]guess-missing-commit-type    If commit type is missing, take a guess about which it is
  --order=feat,fix,test,docs,build,refactor,chore,...
                                      order in which to list commit message types
```

## Semantic Versioning

```text
Usage: changetool semver --from-file=STRING

Manipulate Semantic Versions

Flags:
  -h, --help    Show context-sensitive help.

info
  -d, --debug                Show debugging information
  -l, --log-format="auto"    How to show program output (auto|terminal|jsonl)
  -q, --quiet                Be less verbose than usual

locations
  -p, --path="."               Path for the git worktree/repo to log
  -o, --output="-"             File to which to send output

      --replace-in=FILE,...    Replace version in these files

source
  -n, --max-commits=1000    max number of commits to check
  -s, --since-tag=STRING    Tag from which to start
  -a, --all-commits         report changelog on all commits up to --max-commits. Otherwise, report only to last version tag
      --from-file=STRING    Set previous revision from the first semver looking string found in this file

calculation
  --default-type="fix"                if type is not specified in commit, assume this type
  --[no-]guess-missing-commit-type    If commit type is missing, take a guess about which it is
  --order=feat,fix,test,docs,build,refactor,chore,...
                                      order in which to list commit message types
  --allow-untracked                   allow untracked files to count as clean
```
