package scheduler

import "errors"

var (
	ErrJobNotFound = errors.New("job not found")
	ErrUnsupportedType = errors.New("unsupported server type")
	ErrServerExists    = errors.New("server already exists")
	ErrServerNotFound  = errors.New("server not found")
)
