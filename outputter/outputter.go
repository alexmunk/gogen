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
)

// InitROT starts the Read Out Thread which will log statistics about what's being output
// InitROT is intended to be started as a goroutine which will log output every c.
func InitROT(c *config.Config) {
	log = c.Log

	lastEventsWritten := eventsWritten
	lastBytesWritten := bytesWritten
	var gbday, eventssec, bytessec float64
	var tempEW, tempBW int64
	lastTS = time.Now()
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second)
		<-timer.C
		n := time.Now()
		tempEW = eventsWritten
		tempBW = bytesWritten
		eventssec = float64(tempEW-lastEventsWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval)
		bytessec = float64(tempBW-lastBytesWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval)
		gbday = (bytessec * 60 * 60 * 24) / 1024 / 1024 / 1024
		log.Infof("Events/Sec: %2f Kilobytes/Sec: %2f GB/Day: %2f", eventssec, bytessec, gbday)
		lastTS = n
		lastEventsWritten = tempEW
		lastBytesWritten = tempBW
	}
}

func Start(oq chan *config.OutQueueItem, oqs chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	for {
		item, ok := <-oq
		if !ok {
			oqs <- 1
			break
		}
		setup(generator, item)
		if len(item.Events) > 0 {
			go func() {
				defer item.IO.W.Close()
				// We'll crash on empty events, but don't do that!
				getLine("header", item.S, item.Events[0], item.IO.W)
				// item.S.Log.Debugf("Out Queue Item %#v", item)
				var last int
				for i, line := range item.Events {
					getLine("row", item.S, line, item.IO.W)
					last = i
				}
				getLine("footer", item.S, item.Events[last], item.IO.W)
			}()
		}
		err := item.S.Out.Send(item)
		if err != nil {
			item.S.Log.Errorf("Error with Send(): %s", err)
		}
	}
}

func getLine(templatename string, s *config.Sample, line map[string]string, w io.Writer) error {
	if template.Exists(s.OutputTemplate + "_" + templatename) {
		linestr, err := template.Exec(s.OutputTemplate+"_"+templatename, line)
		if err != nil {
			s.Log.Errorf("Error from sample '%s' in template execution: %v", s.Name, err)
			return err
		}
		// item.S.Log.Debugf("Outputting line %s", linestr)
		_, err = w.Write([]byte(linestr + "\n"))
		if err != nil {
			s.Log.Errorf("Error sending event for sample '%s' to outputter '%s': %s", s.Name, s.Outputter, err)
		}
	}
	return nil
}

func setup(generator *rand.Rand, item *config.OutQueueItem) {
	item.Rand = generator
	item.IO = config.NewOutputIO()
	// Check to see if our outputter is not set
	if item.S.Out == nil {
		item.S.Log.Infof("Setting sample '%s' to outputter '%s'", item.S.Name, item.S.Outputter)
		switch item.S.Outputter {
		case "stdout":
			item.S.Out = new(stdout)
		case "devnull":
			item.S.Out = new(devnull)
		}
	}
}

// // Writer implements io.Writer, but allows for Teeing the data for the ReadOutThread
// type Writer interface {
// 	io.Writer
// }

// // Write implements io.Writer, but allows for Teeing the data for the ReadOutThread
// func (writer Writer) Write(p []byte) (n int, err error) {

// }
