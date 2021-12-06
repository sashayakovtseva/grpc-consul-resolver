package consul

// Option is used to configure Resolver.
type Option func(r *Resolver)

// WithLogger sets logger.
func WithLogger(l Logger) Option {
	return func(r *Resolver) {
		r.logger = l
	}
}
