package rater

import (
	"time"

	config "github.com/coccyx/gogen/internal"
)

// DefaultRater simply returns the passed count
type DefaultRater struct {
	c *config.RaterConfig
}

// GetRate implements Rater interface
func (dr *DefaultRater) GetRate(now time.Time) float64 {
	return 1.0
}

// EventRate takes a given sample and current count and returns the rated count
func (dr *DefaultRater) EventRate(s *config.Sample, now time.Time, count int) int {
	return EventRate(s, now, count)
}

// TokenRate takes a token and returns the rated value
func (dr *DefaultRater) TokenRate(t config.Token, now time.Time) float64 {
	return TokenRate(t, now)
}
