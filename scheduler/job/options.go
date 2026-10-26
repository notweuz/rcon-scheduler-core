package job

import (
	"github.com/rs/zerolog"
)

// Option configures a [Job] created with [New].
type Option func(o *options)

type options struct {
	logger zerolog.Logger
}

// WithLogger sets the logger used by itself.
// By default all log output is discarded.
func WithLogger(l zerolog.Logger) Option {
	return func(o *options) { 
		o.logger = l 
	}
}
