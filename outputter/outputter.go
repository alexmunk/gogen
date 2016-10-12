package outputter

import (
	"io"
	"math/rand"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/template"
	logging "github.com/op/go-logging"
)

var (
	eventsWritten int64
	bytesWritten  int64
	lastTS        time.Time
	log           *logging.Logger
	rotchan       chan *config.OutputStats
	gout          [config.MaxOutputThreads]config.Outputter
)

// ROT starts the Read Out Thread which will log statistics about what's being output
// ROT is intended to be started as a goroutine which will log output every c.
func ROT(c *config.Config) {
	log = c.Log

	rotchan = make(chan *config.OutputStats)
	go readStats()

	lastEventsWritten := eventsWritten
	lastBytesWritten := bytesWritten
	var gbday, eventssec, kbytessec float64
	var tempEW, tempBW int64
	lastTS = time.Now()
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second)
		<-timer.C
		n := time.Now()
		tempEW = eventsWritten
		tempBW = bytesWritten
		eventssec = float64(tempEW-lastEventsWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval)
		kbytessec = float64(tempBW-lastBytesWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval) / 1024
		gbday = (kbytessec * 60 * 60 * 24) / 1024 / 1024
		log.Infof("Events/Sec: %.2f Kilobytes/Sec: %.2f GB/Day: %.2f", eventssec, kbytessec, gbday)
		lastTS = n
		lastEventsWritten = tempEW
		lastBytesWritten = tempBW
	}
}

func readStats() {
	for {
		select {
		case os := <-rotchan:
			eventsWritten += os.EventsWritten
			bytesWritten += os.BytesWritten
		}
	}
}

// Account sends eventsWritten and bytesWritten to the readStats() thread
func Account(eventsWritten int64, bytesWritten int64) {
	os := new(config.OutputStats)
	os.EventsWritten = eventsWritten
	os.BytesWritten = bytesWritten
	rotchan <- os
}

// Start starts an output thread and runs until notified to shut down
func Start(oq chan *config.OutQueueItem, oqs chan int, num int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	var lastS *config.Sample
	var out config.Outputter
	for {
		item, ok := <-oq
		if !ok {
			if lastS != nil {
				lastS.Log.Infof("Closing output for sample '%s'", lastS.Name)
				out.Close()
				gout[num] = nil
			}
			oqs <- 1
			break
		}
		out = setup(generator, item, num)
		if len(item.Events) > 0 {
			go func() {
				defer item.IO.W.Close()
				switch item.S.Output.OutputTemplate {
				case "raw":
					for _, line := range item.Events {
						_, err := item.IO.W.Write([]byte(line["_raw"]))
						_, err = item.IO.W.Write([]byte("\n"))
						if err != nil {
							item.S.Log.Errorf("Error writing to IO Buffer: %s", err)
						}
					}
				default:
					// We'll crash on empty events, but don't do that!
					getLine("header", item.S, item.Events[0], item.IO.W)
					// item.S.Log.Debugf("Out Queue Item %#v", item)
					var last int
					for i, line := range item.Events {
						getLine("row", item.S, line, item.IO.W)
						last = i
					}
					getLine("footer", item.S, item.Events[last], item.IO.W)
				}
			}()
			err := out.Send(item)
			if err != nil {
				item.S.Log.Errorf("Error with Send(): %s", err)
			}
		}
		lastS = item.S
	}
}

func getLine(templatename string, s *config.Sample, line map[string]string, w io.Writer) error {
	if template.Exists(s.Output.OutputTemplate + "_" + templatename) {
		linestr, err := template.Exec(s.Output.OutputTemplate+"_"+templatename, line)
		if err != nil {
			s.Log.Errorf("Error from sample '%s' in template execution: %v", s.Name, err)
			return err
		}
		// item.S.Log.Debugf("Outputting line %s", linestr)
		_, err = w.Write([]byte(linestr))
		_, err = w.Write([]byte("\n"))
		if err != nil {
			s.Log.Errorf("Error sending event for sample '%s' to outputter '%s': %s", s.Name, s.Output.Outputter, err)
		}
	}
	return nil
}

func setup(generator *rand.Rand, item *config.OutQueueItem, num int) config.Outputter {
	item.Rand = generator
	item.IO = config.NewOutputIO()

	if gout[num] == nil {
		item.S.Log.Infof("Setting sample '%s' to outputter '%s'", item.S.Name, item.S.Output.Outputter)
		switch item.S.Output.Outputter {
		case "stdout":
			gout[num] = new(stdout)
		case "devnull":
			gout[num] = new(devnull)
		case "file":
			gout[num] = new(file)
		case "http":
			gout[num] = new(httpout)
		case "buf":
			gout[num] = new(buf)
		default:
			gout[num] = new(stdout)
		}
	}
	return gout[num]
}

// // Writer implements io.Writer, but allows for Teeing the data for the ReadOutThread
// type Writer interface {
// 	io.Writer
// }

// // Write implements io.Writer, but allows for Teeing the data for the ReadOutThread
// func (writer Writer) Write(p []byte) (n int, err error) {

// }
