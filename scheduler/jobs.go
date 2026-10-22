package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/notweuz/rcon-scheduler-core/scheduler/job"
	"github.com/notweuz/rcon-scheduler-core/server"
)

// AddJob schedules j on the server identified by key and returns the job ID,
// which can be passed to [Scheduler.RemoveJob].
//
// The job is validated with [job.Job.Validate] before scheduling.
// A job never overlaps with itself: if a run is still in progress when the next
// one is due, the next run is skipped.
//
// It returns [ErrServerNotFound] if the server is not registered.
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

// RemoveJob removes the job with the given id from the schedule.
//
// If the job is running at that moment, its context is cancelled and it stops
// before its next step. RemoveJob does not wait for the running job to finish.
//
// It returns [ErrJobNotFound] if no job with this id exists.
func (s *Scheduler) RemoveJob(id uuid.UUID) error {
	if err := s.cron.RemoveJob(id); err != nil {
		if errors.Is(err, gocron.ErrJobNotFound) {
			return fmt.Errorf("remove job %s: %w", id, ErrJobNotFound)
		}
		return fmt.Errorf("remove job %s: %w", id, err)
	}

	s.logger.Info().Str("job_id", id.String()).Msg("job removed")
	return nil
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
