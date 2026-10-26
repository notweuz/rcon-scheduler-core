// Package job describes scheduled tasks: a cron expression and a sequence of
// RCON commands and pauses.
package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// ErrNoSteps is returned by [Job.Validate] when the job has no steps.
var ErrNoSteps = errors.New("no steps")

// Job is a named sequence of steps executed on a schedule.
type Job struct {
	// Name identifies the job in logs.
	Name string

	// Cron is a standard 5-field cron expression, e.g. "55 3 * * *".
	// Descriptors such as "@daily" or "@every 1h" are also accepted.
	// It is interpreted in the scheduler's time zone.
	Cron string

	// Steps are executed in order. Execution stops at the first failed step.
	Steps []Step

	logResults bool
}

// New creates a new [Job]. It doesn't run until [Job.Run] is called.
//
// By default server responses are logged at debug level; see [WithLogResults].
func New(name, cron string, steps []Step, opts ...Option) Job {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	return Job{
		Name:       name,
		Cron:       cron,
		Steps:      steps,
		logResults: o.logResults,
	}
}

// Run executes the steps of j in order on c, without any schedule.
//
// Run logs to the logger stored in ctx (see [zerolog.Ctx]); if ctx carries no
// logger, log output is discarded. Each command and its server response is
// logged at debug level, or at info level if j was created with [WithLogResults].
//
// It stops at the first failed command and returns its error.
// If ctx is cancelled during a pause, Run returns ctx.Err() immediately;
// a command already in progress is allowed to finish.
func (j Job) Run(ctx context.Context, c rcon.Client) error {
	log := zerolog.Ctx(ctx)
	for i, step := range j.Steps {
		switch {
		case step.Command != "":
			response, err := c.Execute(ctx, step.Command)
			if err != nil {
				return fmt.Errorf("step %d (%q): %w", i, step.Command, err)
			}
			ev := log.Debug()
			if j.logResults {
				ev = log.Info()
			}
			ev.Int("step", i).Str("command", step.Command).Str("response", response).Msg("command executed")
		case step.Wait > 0:
			log.Debug().Int("step", i).Dur("wait", step.Wait).Msg("waiting")
			select {
			case <-time.After(step.Wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// Validate checks that the cron expression is valid, that j has at least one
// step and that every step is valid (see [Step.Validate]).
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
