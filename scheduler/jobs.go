package scheduler

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/notweuz/rcon-scheduler-core/scheduler/job"
	"github.com/notweuz/rcon-scheduler-core/server"
)

func (s *Scheduler) AddJob(key server.Key, j job.Job) (uuid.UUID, error) {
	if err := j.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("add job %q: %w", j.Name, err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.clients[key]; !ok {
		return uuid.Nil, fmt.Errorf("add job %q: %w", j.Name, ErrServerNotFound)
	}

	gj, err := s.cron.NewJob(
		gocron.CronJob(j.Cron, false),
		gocron.NewTask(s.runJob, key, j),
		gocron.WithName(j.Name),
		gocron.WithTags(key.String()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("add job %q: %w", j.Name, err)
	}

	s.logger.Info().Str("server", key.String()).Str("job", j.Name).Str("cron", j.Cron).Msg("job added")
	return gj.ID(), nil
}

func (s *Scheduler) RemoveJob(id uuid.UUID) error {
	return s.cron.RemoveJob(id)
}

func (s *Scheduler) runJob(ctx context.Context, key server.Key, j job.Job) {
	l := s.logger.With().Str("server", key.String()).Str("job", j.Name).Logger()

	s.mu.RLock()
	client, ok := s.clients[key]
	s.mu.RUnlock()

	if !ok {
		l.Error().Msg("job skipped: server not found")
		return
	}

	l.Info().Msg("job started")
	if err := j.Run(ctx, client); err != nil {
		l.Error().Err(err).Msg("job failed")
		return
	}
	l.Info().Msg("job finished")
}