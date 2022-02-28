package main

import (
	"fmt"
	"github.com/deweysasser/golang-program/program"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {

	var options program.Options

	context, err := options.Parse(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// This ends up calling options.Run()
	if err := context.Run(); err != nil {
		log.Err(err).Msg("Program failed")
		os.Exit(1)
	}
}
