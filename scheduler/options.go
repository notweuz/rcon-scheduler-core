package scheduler

import (
	"time"

	"github.com/rs/zerolog"
)

type Option func(*options)

type options struct {
	location *time.Location
	logger   zerolog.Logger
}

func WithLogger(logger zerolog.Logger) Option {
	return func(o *options) { o.logger = logger }
}

func WithLocation(location *time.Location) Option {
	return func(o *options) {
		if location != nil {
			o.location = location
		}
	}
}
