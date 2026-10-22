package job

import (
	"errors"
	"time"
)

// ErrInvalidStep is returned by [Step.Validate] when a step has both or neither
// of Command and Wait set.
var ErrInvalidStep = errors.New("invalid step")

// Step is a single action of a [Job]. Exactly one of Command or Wait must be set.
type Step struct {
	// Command is an RCON command to execute.
	Command string

	// Wait is a pause before the next step.
	Wait time.Duration
}

// Validate returns [ErrInvalidStep] unless exactly one of Command or Wait is set.
func (s Step) Validate() error {
	if (s.Command != "") == (s.Wait > 0) {
		return ErrInvalidStep
	}
	return nil
}
