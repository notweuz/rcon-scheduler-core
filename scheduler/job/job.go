package job

import (
	"context"
	"fmt"
	"time"

	"github.com/notweuz/rcon-scheduler-core/rcon"
)

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
