package generator

import (
	"time"

	"github.com/coccyx/gogen/config"
)

// GenQueueItem represents one generation job
type GenQueueItem struct {
	S        *config.Sample
	Count    int
	Earliest time.Time
	Latest   time.Time
	OQ       chan []string
}

// Generator will generate count events from earliest to latest time and put them
// in the output queue
type Generator interface {
	Gen(item *GenQueueItem) error
}

func Start(gq chan *GenQueueItem) {
	for {
		item := <-gq
		if item.S.Generator == "sample" {
			sampleGen(item)
		}
	}
}

func sampleGen(item *GenQueueItem) {
	c := config.NewConfig()
	c.Log.Debugf("Gen Queue Item %#v", item)
}
