package timer

import (
	"time"

	"github.com/coccyx/gogen/config"
	"github.com/coccyx/gogen/generator"
)

// Timer will put work into the generator queue on an interval specified by the Sample.
// One instance is created per sample.
type Timer struct {
	S  *config.Sample
	GQ chan *generator.GenQueueItem
	OQ chan []string
}

// NewTimer creates a new Timer for a sample which will put work into the generator queue on each interval
func (t *Timer) NewTimer() {
	for {
		timer := time.NewTimer(time.Duration(t.S.Interval) * time.Second)
		<-timer.C
		t.genWork()
	}

}

func (t *Timer) genWork() {
	// c := config.NewConfig()
	// TODO Implement backfill & rating
	earliest := t.S.Now().Add(t.S.EarliestParsed)
	latest := t.S.Now().Add(t.S.LatestParsed)
	item := &generator.GenQueueItem{S: t.S, Count: t.S.Count, Earliest: earliest, Latest: latest, OQ: t.OQ}
	// c.Log.Debugf("Generating work for item %#v", item)
	t.GQ <- item
}
