package cache

import "time"

// Options controls Cache behavior.
type Options struct {
	// Capacity is the maximum number of live entries the cache will keep.
	// When the capacity is exceeded, least recently used items are evicted.
	// Must be > 0. Defaults to 1024.
	Capacity int

	// CleanupInterval is used by the background janitor when there are no
	// scheduled expirations to wait for. Defaults to 1 minute.
	CleanupInterval time.Duration

	// EnableJanitor starts a lightweight background goroutine that removes
	// expired items to free memory proactively. Defaults to true.
	EnableJanitor bool
}

// Option mutates Options.
type Option func(*Options)

// WithCapacity sets the maximum number of entries.
func WithCapacity(n int) Option {
	return func(o *Options) { o.Capacity = n }
}

// WithCleanupInterval sets the background cleanup fallback interval.
func WithCleanupInterval(d time.Duration) Option {
	return func(o *Options) { o.CleanupInterval = d }
}

// WithJanitor enables/disables the background janitor.
func WithJanitor(enable bool) Option {
	return func(o *Options) { o.EnableJanitor = enable }
}

func defaultOptions() Options {
	return Options{
		Capacity:        1024,
		CleanupInterval: time.Minute,
		EnableJanitor:   true,
	}
}
