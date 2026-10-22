package scheduler

import "errors"

var (
	// ErrJobNotFound is returned by [Scheduler.RemoveJob] when no job with the given id exists.
	ErrJobNotFound = errors.New("job not found")

	// ErrUnsupportedType is returned by [Scheduler.Register] when there is no
	// client implementation for the server type.
	ErrUnsupportedType = errors.New("unsupported server type")

	// ErrServerExists is returned by [Scheduler.Register] when a server with the
	// same key is already registered.
	ErrServerExists = errors.New("server already exists")

	// ErrServerNotFound is returned when no server with the given key is registered.
	ErrServerNotFound = errors.New("server not found")
)
