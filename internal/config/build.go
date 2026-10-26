package config

import (
	"fmt"
	"time"

	"github.com/notweuz/rcon-scheduler-core/scheduler/job"
	"github.com/notweuz/rcon-scheduler-core/server"
)

type Entry struct {
	Server server.Server
	Jobs []job.Job
}

func (c *Config) Build() ([]Entry, error) {
	entries := make([]Entry, 0, len(c.Servers))
	for i, s := range c.Servers {
		jobs := make([]job.Job, 0, len(s.Jobs))
		for k, j := range s.Jobs {
			jb, err := j.toJob()
			if err != nil {
				return nil, fmt.Errorf("servers[%d].jobs[%d]: %w", i, k, err)
			}
			jobs = append(jobs, jb)
		}
		entries = append(entries, Entry{
			Server: server.Server{
				Address:  s.Address,
				Port:     s.Port,
				Password: s.Password,
				Type:     s.Type,
			},
			Jobs: jobs,
		})
	}
	return entries, nil
}

func (j Job) toJob() (job.Job, error) {
	steps := make([]job.Step, 0, len(j.Steps))
	for i, s := range j.Steps {
		var st job.Step
		if s.Execute != nil {
			st.Command = *s.Execute
		}
		if s.Wait != nil {
			d, err := time.ParseDuration(*s.Wait)
			if err != nil {
				return job.Job{}, fmt.Errorf("steps[%d]: invalid wait %q: %w", i, *s.Wait, err)
			}
			st.Wait = d
		}
		steps = append(steps, st)
	}
	return job.New(j.Name, j.Cron, steps), nil
}