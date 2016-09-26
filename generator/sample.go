package generator

import (
	"math"

	"github.com/coccyx/gogen/internal"
)

type sample struct{}

func (foo sample) Gen(item *config.GenQueueItem) error {
	s := item.S
	if item.Count == -1 {
		item.Count = len(s.Lines)
	}
	s.Log.Debugf("Gen Queue Item %#v", item)
	// outstr := []map[string]string{{"_raw": fmt.Sprintf("%#v", item)}}

	s.Log.Debugf("Generating sample '%s' with count %d, et: '%s', lt: '%s'", s.Name, item.Count, item.Earliest, item.Latest)
	// startTime := time.Now()

	slen := len(s.Lines)

	// Make a template of events we'll be generating every interval, by randomly sampling from the sample
	// or filling in sequentially.  When done, cache the output and copy that cached output rather
	// than reallocate every interval.
	var events []map[string]string
	if s.Events == nil || len(s.Events) != item.Count {
		events = make([]map[string]string, 0, item.Count)
		if s.RandomizeEvents {
			s.Log.Debugf("Random filling events for sample '%s' with %d events", s.Name, item.Count)

			for i := 0; i < item.Count; i++ {
				events = append(events, copyevent(s.Lines[item.Rand.Intn(slen)]))
			}
		} else {
			if item.Count <= slen {
				for i := 0; i < item.Count; i++ {
					events = append(events, copyevent(s.Lines[i]))
				}
			} else {
				iters := int(math.Ceil(float64(item.S.Count) / float64(slen)))
				s.Log.Debugf("Sequentially filling events for sample '%s' of size %d with %d events over %d iterations", s.Name, slen, item.Count, iters)
				for i := 0; i < iters; i++ {
					var count int
					// start := i * slen
					if i == iters-1 {
						count = (item.S.Count - (i * slen))
					} else {
						count = slen
					}
					s.Log.Debugf("Appending %d events from lines, length %d", count, slen)
					// end := (i * slen) + count
					for j := 0; j < count; j++ {
						events = append(events, copyevent(s.Lines[j]))
					}
				}
			}
		}
		copy(s.Events, events)
	} else {
		copy(events, s.Events)
	}

	s.Log.Debugf("Events: %#v", events)

	for i := 0; i < item.Count; i++ {
		choices := make(map[int]*int)
		for _, token := range s.Tokens {
			if fieldval, ok := events[i][token.Field]; ok {
				var choice *int
				if choices[token.Group] != nil {
					choice = choices[token.Group]
				} else {
					choice = new(int)
					*choice = -1
				}
				s.Log.Debugf("Replacing token '%s':'%s' with choice %d in fieldval: %s", token.Name, token.Token, *choice, fieldval)
				if err := token.Replace(&fieldval, choice, item.Earliest, item.Latest, item.Rand); err == nil {
					events[i][token.Field] = fieldval
				} else {
					s.Log.Error(err)
				}
				if token.Group > 0 {
					choices[token.Group] = choice
				}
			} else {
				s.Log.Errorf("Field %s not found in event for sample %s", token.Field, s.Name)
			}
		}
	}

	outitem := &config.OutQueueItem{S: item.S, Events: events}
	item.OQ <- outitem
	return nil
}

func copyevent(src map[string]string) (dst map[string]string) {
	dst = make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
