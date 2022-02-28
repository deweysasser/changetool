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

		c, err := program.Parse([]string{"version"})

		assert.NoError(t, err)

		c.Run(&program)
	})

	// VersionCmd does not explicitly call os.Exit
	assert.Equal(t, -1, exitValue)
	assert.Equal(t, "unknown\n", out)
}
