package job

// Option configures a [Job] created with [New].
type Option func(o *options)

type options struct {
	logResults bool
}

// WithLogResults makes [Job.Run] log each command and its server response at
// info level instead of debug level.
func WithLogResults(enabled bool) Option {
	return func(o *options) { o.logResults = enabled }
}
