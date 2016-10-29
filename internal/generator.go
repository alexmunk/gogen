package config

import (
	"math/rand"
	"time"
)

// GenQueueItem represents one generation job
type GenQueueItem struct {
	S        *Sample
	Count    int
	Earliest time.Time
	Latest   time.Time
	Now      time.Time
	OQ       chan *OutQueueItem
	Rand     *rand.Rand
}

// Generator will generate count events from earliest to latest time and put them
// in the output queue
type Generator interface {
	Gen(item *GenQueueItem) error
}
