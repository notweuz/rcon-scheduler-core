package job

import (
	"context"
	"fmt"
	"time"

	"github.com/notweuz/rcon-scheduler-core/rcon"

	"github.com/rs/zerolog/log"
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
			resp, err := c.Execute(ctx, step.Command)
			if err != nil {
				return fmt.Errorf("step %d (%q): %w", i, step.Command, err)
			}
			log.Debug().Str("cmd", step.Command).Str("resp", resp).Msg("executed")
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
