package scheduler

import "github.com/go-co-op/gocron/v2"

type Scheduler struct {
	cron gocron.Scheduler
}
