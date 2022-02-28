package program

import (
	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
	"os"
	"testing"
)

func TestOptions_Run(t *testing.T) {
	var program Options

	exitValue := -1
	fakeExit := func(x int) {
		exitValue = x
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	out := capturer.CaptureStdout(func() {

		_, err := program.Parse([]string{"--version"})

		assert.NoError(t, err)

		// version output is done as part of parsing, so we don't need to run the program
	})

	assert.Equal(t, exitValue, 0)
	assert.Equal(t, "unknown\n", out)
}
