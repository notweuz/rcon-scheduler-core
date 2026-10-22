package job

import "time"

type Step struct {
	Command string
	Wait    time.Duration
}
