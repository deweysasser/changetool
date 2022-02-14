package main

import (
	"github.com/alecthomas/kong"
	"github.com/deweysasser/concom/program"
	"github.com/rs/zerolog/log"
)

func main() {

	var Options program.Options

	context := kong.Parse(&Options,
		kong.Description("Brief Program Summary"),
		kong.Vars{
			"type_order": program.TypesInOrder.Join(","),
		},
	)

	Options.Init()

	// This ends up calling Options.Run()
	if err := context.Run(); err != nil {
		log.Err(err).Msg("Program failed")
	}
}
