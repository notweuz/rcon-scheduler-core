package scheduler

import (
	"time"

	"github.com/rs/zerolog"
)

// Option configures a [Scheduler] created with [New].
type Option func(*options)

type options struct {
	location *time.Location
	logger   zerolog.Logger
}

// WithLogger sets the logger used by the scheduler and its RCON clients.
// Each client gets a child logger with a "server" field.
// By default all log output is discarded.
func WithLogger(logger zerolog.Logger) Option {
	return func(o *options) { o.logger = logger }
}

// WithLocation sets the time zone in which cron expressions are interpreted.
// A nil location is ignored. The default is UTC.
func WithLocation(location *time.Location) Option {
	return func(o *options) {
		if location != nil {
			o.location = location
		}
	}
}
