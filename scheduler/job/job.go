package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/robfig/cron/v3"
)

var ErrNoSteps = errors.New("no steps")

type Job struct {
	Name  string
	Cron  string
	Steps []Step
}

func (j Job) Run(ctx context.Context, c rcon.Client) error {
	for i, step := range j.Steps {
		switch {
		case step.Command != "":
			_, err := c.Execute(ctx, step.Command)
			if err != nil {
				return fmt.Errorf("step %d (%q): %w", i, step.Command, err)
			}
		case step.Wait > 0:
			select {
			case <-time.After(step.Wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

func (j Job) Validate() error {
	if _, err := cron.ParseStandard(j.Cron); err != nil {
		return fmt.Errorf("invalid cron %q: %w", j.Cron, err)
	}
	if len(j.Steps) == 0 {
		return ErrNoSteps
	}
	for i, s := range j.Steps {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("step[%d]: %w", i, err)
		}
	}
	return nil
}