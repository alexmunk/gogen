package timer

import (
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/rater"
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
		log.Infof("Timer for sample '%s' shutting down after %d intervals", t.S.Name, t.S.EndIntervals)
		t.Done <- 1
	} else {
		if !t.S.Realtime {
			var endtime time.Time
			n := time.Now()
			if t.S.EndParsed.Before(n) && !t.S.EndParsed.IsZero() {
				endtime = t.S.EndParsed
			} else {
				endtime = n
			}
			for ; t.S.Current.Before(endtime); t.S.Current = t.S.Current.Add(time.Duration(t.S.Interval) * time.Second) {
				// log.Debugf("Backfilling, at %s, ending at %s", t.S.Current, endtime)
				t.genWork()
			}
			if t.S.EndParsed.IsZero() {
				t.S.Realtime = true
			}
		}
		// We'll be greater than now but we still need to continue to the end
		if !t.S.Realtime {
			for ; t.S.Current.Before(t.S.EndParsed); t.S.Current = t.S.Current.Add(time.Duration(t.S.Interval) * time.Second) {
				t.genWork()
			}
		}
		if t.S.Realtime {
			for {
				timer := time.NewTimer(time.Duration(t.S.Interval) * time.Second)
				<-timer.C
				t.genWork()
			}
		} else {
			t.Done <- 1
		}
	}
}

func (t *Timer) genWork() {
	now := t.S.Now()
	earliest := now.Add(t.S.EarliestParsed)
	latest := now.Add(t.S.LatestParsed)
	count := rater.EventRate(t.S, now, t.S.Count)
	item := &config.GenQueueItem{S: t.S, Count: count, Earliest: earliest, Latest: latest, Now: now, OQ: t.OQ}
	// log.Debugf("Placing item in queue for sample '%s': %#v", t.S.Name, item)
	t.GQ <- item
}
