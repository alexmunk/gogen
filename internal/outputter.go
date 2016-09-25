package config

import "math/rand"

// OutQueueItem represents one batch of events to output
type OutQueueItem struct {
	S      *Sample
	Events []map[string]string
	Rand   *rand.Rand
}

// Outputter will output events using the designated output plugin
type Outputter interface {
	Send(item *OutQueueItem) error
}
