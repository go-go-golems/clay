package builder

// Options contains configuration for the query builder
type Options struct {
	// DefaultFieldBoost is the default boost value for field queries
	DefaultFieldBoost float64
	// DefaultConjunctionMinimum is the minimum number of clauses that must match for conjunction queries
	DefaultConjunctionMinimum float64
	// DefaultDisjunctionMinimum is the minimum number of clauses that must match for disjunction queries
	DefaultDisjunctionMinimum float64
}

// Option is a function that configures the builder options
type Option func(*Options)

// DefaultOptions returns the default options
func DefaultOptions() *Options {
	return &Options{
		DefaultFieldBoost:         1.0,
		DefaultConjunctionMinimum: 1.0,
		DefaultDisjunctionMinimum: 1.0,
	}
}

// WithDefaultFieldBoost sets the default boost value for field queries
func WithDefaultFieldBoost(boost float64) Option {
	return func(o *Options) {
		o.DefaultFieldBoost = boost
	}
}

// WithDefaultConjunctionMinimum sets the minimum number of clauses that must match for conjunction queries
func WithDefaultConjunctionMinimum(min_ float64) Option {
	return func(o *Options) {
		o.DefaultConjunctionMinimum = min_
	}
}

// WithDefaultDisjunctionMinimum sets the minimum number of clauses that must match for disjunction queries
func WithDefaultDisjunctionMinimum(min_ float64) Option {
	return func(o *Options) {
		o.DefaultDisjunctionMinimum = min_
	}
}
