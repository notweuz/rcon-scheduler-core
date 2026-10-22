package job

import (
	"errors"
	"time"
)

var ErrInvalidStep = errors.New("invalid step")

type Step struct {
	Command string
	Wait    time.Duration
}

func (s Step) Validate() error {
	if (s.Command != "") == (s.Wait > 0) {
		return ErrInvalidStep
	}
	return nil
}