package config

import (
	"time"
)

type Sample struct {
	Name            string              `json:"name"`
	Disabled        bool                `json:"disabled"`
	Generator       string              `json:"generator"`
	Outputter       string              `json:"outputter"`
	Rater           string              `json:"rater"`
	Interval        int                 `json:"interval"`
	Delay           int                 `json:"delay"`
	Count           int                 `json:"count"`
	Earliest        string              `json:"earliest"`
	Latest          string              `json:"latest"`
	Begin           string              `json:"begin"`
	End             string              `json:"end"`
	RandomizeCount  float32             `json:"randomizeCount"`
	RandomizeEvents bool                `json:"randomizeEvents"`
	Tokens          []Token             `json:"tokens"`
	Lines           []string            `json:"lines"`
	LinesMap        []map[string]string `json:"linesMap"`

	// Internal use variables
	earliestParsed time.Duration `json:"-"`
	latestParsed   time.Duration `json:"-"`
	beginParsed    time.Time     `json:"-"`
	endParsed      time.Time     `json:"-"`
	current        time.Time     `json:"-"` // If we are backfilling or generating for a specified time window, what time is it?
	realtime       bool          `json:"-"` // Are we done doing batch backfill or specified time window?
}

// Clock allows for implementers to keep track of their own view
// of current time.  In Gogen, this is used for being able to generate
// events between certain time windows, or backfill from a certain time
// while continuing to run in real time.
type Clock interface {
	Now() time.Time
}

// Now returns the current time for the sample, and handles
// whether we are currently generating a backfill or
// specified time window or whether we should be generating
// events in realtime
func (s *Sample) Now() time.Time {
	if !s.realtime {
		return s.current
	} else {
		return time.Now()
	}
}

type Token struct {
	Name        string                       `json:"name"`
	Format      string                       `json:"format"`
	Token       string                       `json:"token"`
	Type        string                       `json:"type"`
	Replacement string                       `json:"replacement"`
	Sample      string                       `json:"sample"`
	Field       string                       `json:"field"`
	Precision   int                          `json:"precision"`
	Lower       int                          `json:"lower"`
	Upper       int                          `json:"upper"`
	PercChoice  map[string]float32           `json:"percEnum"`
	FieldChoice map[string]map[string]string `json:"fieldNum"`
	Choice      []string                     `json:"choice"`
}
