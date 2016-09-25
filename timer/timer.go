package timer

import (
	"time"

	"github.com/coccyx/gogen/internal"
)

// Timer will put work into the generator queue on an interval specified by the Sample.
// One instance is created per sample.
type Timer struct {
	S    *config.Sample
	GQ   chan *config.GenQueueItem
	OQ   chan *config.OutQueueItem
	Done chan int
}

// NewTimer creates a new Timer for a sample which will put work into the generator queue on each interval
func (t *Timer) NewTimer() {
	if t.S.EndIntervals > 0 {
		for i := 0; i < t.S.EndIntervals; i++ {
			t.genWork()
		}
		t.S.Log.Infof("Timer for sample '%s' shutting down after %d intervals", t.S.Name, t.S.EndIntervals)
		t.Done <- 1
	} else {
		for {
			t.loop()
		}
	}
}

func (t *Timer) loop() {
	timer := time.NewTimer(time.Duration(t.S.Interval) * time.Second)
	<-timer.C
	t.genWork()
}

func (t *Timer) genWork() {
	// TODO Implement backfill & rating
	earliest := t.S.Now().Add(t.S.EarliestParsed)
	latest := t.S.Now().Add(t.S.LatestParsed)
	item := &config.GenQueueItem{S: t.S, Count: t.S.Count, Earliest: earliest, Latest: latest, OQ: t.OQ}
	t.S.Log.Debugf("Placing item in queue for sample '%s': %#v", t.S.Name, item)
	t.GQ <- item
	// select {
	// case t.GQ <- item:
	// default:
	// }
}
