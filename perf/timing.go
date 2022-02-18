package perf

import (
	"github.com/rs/zerolog/log"
	"time"
)

type TimeLogger struct {
	Name  string
	Start time.Time
}

func (t TimeLogger) Stop() {
	log.Debug().
		Str("timer", t.Name).
		Dur("duration", time.Since(t.Start)).
		Msg("timer finished")
}

func Timer(name string) TimeLogger {
	now := time.Now()
	log.Debug().
		Str("timer", name).
		Time("start", now).
		Msg("starting timer")
	return TimeLogger{
		Name:  name,
		Start: now,
	}
}
