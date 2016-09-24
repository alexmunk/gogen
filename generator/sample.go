package generator

import (
	"math"
	"math/rand"

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
	events := make([]map[string]string, 0, item.Count)

	if s.RandomizeEvents {
		s.Log.Debugf("Random filling events for sample '%s' with %d events", s.Name, item.Count)

		for i := 0; i < item.Count; i++ {
			events = append(events, s.Lines[rand.Intn(slen)])
		}
	} else {
		if item.Count <= slen {
			copy(events, s.Lines[0:item.Count])
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
				events = append(events, s.Lines[0:count]...)
			}
		}
	}

	s.Log.Debugf("Events: %#v", events)

	choices := make(map[string]*int)
	for i := 0; i < item.Count; i++ {
		for _, token := range s.Tokens {
			if fieldval, ok := events[i][token.Field]; ok {
				s.Log.Debugf("Replacing token '%s' in fieldval: %s", token.Name, fieldval)
				if choices[token.Name] == nil {
					choices[token.Name] = new(int)
				}
				if err := token.Replace(&fieldval, choices[token.Name], item.Earliest, item.Latest); err == nil {
					events[i][token.Field] = fieldval
				} else {
					s.Log.Error(err)
				}
			} else {
				s.Log.Errorf("Field %s not found in event for sample %s", token.Field, s.Name)
			}
		}
	}

	// s.Log.Debugf("Outstr: ", outstr)
	outitem := &config.OutQueueItem{S: item.S, Events: events}
	select {
	case item.OQ <- outitem:
	default:
	}
	return nil
}
