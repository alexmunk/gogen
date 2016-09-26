package outputter

import (
	"math/rand"
	"time"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/template"
)

func Start(oq chan *config.OutQueueItem, oqs chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	for {
		item, ok := <-oq
		if !ok {
			oqs <- 1
			break
		}
		item.Rand = generator
		SetOutputter(item.S)
		event := ""
		if len(item.Events) > 0 {
			// We'll crash on empty events, but don't do that!
			event += getLine("header", item.S, item.Events[0])
			// item.S.Log.Debugf("Out Queue Item %#v", item)
			var last int
			for i, line := range item.Events {
				event += getLine("row", item.S, line)
				last = i
			}
			event += getLine("footer", item.S, item.Events[last])
		}
		err := item.S.Out.Send(event)
		if err != nil {
			item.S.Log.Errorf("Error sending event for sample '%s' to outputter '%s'", item.S.Name, item.S.Outputter)
		}
	}
}

func getLine(templatename string, s *config.Sample, line map[string]string) string {
	if template.Exists(s.OutputTemplate + "_" + templatename) {
		linestr, err := template.Exec(s.OutputTemplate+"_"+templatename, line)
		if err != nil {
			s.Log.Errorf("Error from sample '%s' in template execution: %v", s.Name, err)
			return ""
		}
		// item.S.Log.Debugf("Outputting line %s", linestr)
		return linestr
	}
	return ""
}

func SetOutputter(s *config.Sample) {
	// Check to see if our outputter is not set
	if s.Out == nil {
		s.Log.Infof("Setting sample '%s' to outputter '%s'", s.Name, s.Outputter)
		switch s.Outputter {
		case "stdout":
			o := new(stdout)
			s.Out = o
		case "devnull":
			o := new(devnull)
			s.Out = o
		}
	}
}
