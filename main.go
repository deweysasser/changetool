package main

import (
	"github.com/deweysasser/changetool/program"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {

	var options program.Options

	context, err := options.Parse(os.Args[1:])

	if err != nil {
		os.Exit(1)
	}

	// This ends up calling options.Run()
	if err := context.Run(); err != nil {
		log.Err(err).Msg("Program failed")
		os.Exit(1)
	}
}
