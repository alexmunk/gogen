package config

import (
	"math/rand"
	"time"
)

// GeneratorConfig holds our configuration for custom generators
type GeneratorConfig struct {
	Name           string                 `json:"name"`
	Init           map[string]string      `json:"init,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
	Script         string                 `json:"script"`
	FileName       string                 `json:"fileName,omitempty"`
	SingleThreaded bool                   `json:"singleThreaded,omitempty"`
}

// GenQueueItem represents one generation job
type GenQueueItem struct {
	S        *Sample
	Count    int
	Event    int
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
