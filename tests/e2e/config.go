package e2e

// Test configuration constants
const (
	// Default broker connection URL
	DefaultBrokerURL = "amqp://guest:guest@localhost:5672/"

	// Test timeouts
	DefaultTimeout = 2000 // milliseconds
	ShortTimeout   = 500  // milliseconds
	LongTimeout    = 5000 // milliseconds

	// Test prefetch values
	SmallPrefetch     = 2
	MediumPrefetch    = 5
	LargePrefetch     = 20
	UnlimitedPrefetch = 0
)
