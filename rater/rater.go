package rater

import (
	"reflect"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

// EventRate takes a given sample and current count and returns the rated count
func EventRate(s *config.Sample, now time.Time, count int) (ret int) {
	if s.Rater == nil {
		s.Rater = GetRater(s.RaterString)
		log.Infof("Setting rater to type %s, for sample '%s'", reflect.TypeOf(s.Rater), s.Name)
	}
	rate := s.Rater.GetRate(now)
	ratedCount := rate * float64(count)
	if ratedCount < 0 {
		ret = int(ratedCount - 0.5)
	} else {
		ret = int(ratedCount + 0.5)
	}
	return ret
}

// TokenRate takes a token and returns the rated value
func TokenRate(t *config.Token, now time.Time) float64 {
	if t.Rater == nil {
		t.Rater = GetRater(t.RaterString)
		log.Infof("Setting rater to type %s, for token '%s'", reflect.TypeOf(t.Rater), t.Name)
	}
	return t.Rater.GetRate(now)
}

// GetRater returns a rater interface
func GetRater(name string) (ret config.Rater) {
	c := config.NewConfig()
	r := c.FindRater(name)
	if r == nil {
		r := c.FindRater("default")
		ret = &DefaultRater{c: r}
	} else if r.Name == "default" {
		r := c.FindRater("default")
		ret = &DefaultRater{c: r}
	} else if r.Type == "config" {
		ret = &ConfigRater{c: r}
	} else {
		ret = &ScriptRater{c: r}
	}
	return ret
}
