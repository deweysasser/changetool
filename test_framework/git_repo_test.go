package test_framework

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BuildGitRepo(t *testing.T) {
	repo, err := New("../output/test_framework/Test_BuildGitRepo")

	if err != nil {
		assert.FailNow(t, err.Error(), "Failure initializing repo")
	}

	err = repo.RunFile("git_repo_test.yaml")

	if err != nil {
		assert.FailNow(t, err.Error(), "Failed to read YAML")
	}

	if err != nil {
		assert.FailNow(t, err.Error(), "Failure initializing repo")
	}
}
