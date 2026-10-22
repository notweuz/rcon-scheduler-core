package scheduler

import (
	"log/slog"
	"time"
)

type Option func(*options)

type options struct {
	location *time.Location
	logger *slog.Logger
}

func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func WithLocation(location *time.Location) Option {
	return func(o *options) {
		if location != nil {
			o.location = location
		}
	}
}