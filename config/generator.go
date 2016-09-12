package config

import "time"

// GenQueueItem represents one generation job
type GenQueueItem struct {
	S        *Sample
	Count    int
	Earliest time.Time
	Latest   time.Time
	OQ       chan *OutQueueItem
}

// Generator will generate count events from earliest to latest time and put them
// in the output queue
type Generator interface {
	Gen(item *GenQueueItem) error
}
