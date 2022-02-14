# concom

A tool for working with projects that use the [conventional commit](https://www.conventionalcommits.org/) conventions.

## Status

Experimental.  This seems to work, but I haven't put in good test cases.

## Overview

This is a first, very simple, pass at a tool for working with conventional commits projects.

The other tools that I've found are unsatisfactory in some way, either by having *way* too many
dependencies (I'm looking gat you, Node based tools) or by producing output that I don't really want,
or by requiring templates and config in the project.

## Changelog generation

```text
Usage: concom changelog

Flags:
  -h, --help                              Show context-sensitive help.
  -d, --debug                             Show debugging information
  -l, --log-format="auto"                 How to show program output (auto|terminal|jsonl)
  -q, --quiet                             Be less verbose than usual

  -p, --path="."                          Path for the git worktree/repo to log
  -s, --since-tag=STRING                  Tag from which to start
      --default-type="fix"                if type is not specified in commit, assume this type
      --[no-]guess-missing-commit-type    If commit type is missing, take a guess about which it is
      --order=feat,fix,test,docs,build,refactor,chore,...
                                          order in which to list commit message types
  ```

## Semantic Versioning

```text
Usage: concom semver --from-tag --from-file=STRING

Manipulate Semantic Versions

Flags:
  -h, --help                              Show context-sensitive help.
  -d, --debug                             Show debugging information
  -l, --log-format="auto"                 How to show program output (auto|terminal|jsonl)
  -q, --quiet                             Be less verbose than usual

  -p, --path="."                          Path for the git worktree/repo to log
  -s, --since-tag=STRING                  Tag from which to start
      --default-type="fix"                if type is not specified in commit, assume this type
      --[no-]guess-missing-commit-type    If commit type is missing, take a guess about which it is
      --order=feat,fix,test,docs,build,refactor,chore,...
                                          order in which to list commit message types
      --replace-in=REPLACE-IN,...         Replace version in these files
      --allow-untracked                   allow untracked files to count as clean

source
  --from-tag            Set semver from the last tag
  --from-file=STRING    Set previous revision from the first semver looking string found in this file
```

## Examples

Generate a changelog since version 1.2

```text
concom changelog --since-tag v1.2
```

Find out the next version
```text
concom semver --from-tag
```

Replace a file with the next version tag
```text
concom semver --from-tag --replace-in version.txt
```