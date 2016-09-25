package outputter

import (
	"math/rand"
	"time"

	"github.com/coccyx/gogen/internal"
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
		item.S.Out.Send(item)
	}
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
