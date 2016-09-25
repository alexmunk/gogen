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
		// Check to see if our outputter is not set
		if item.S.Out == nil {
			item.S.Log.Infof("Setting sample '%s' to outputter '%s'", item.S.Name, item.S.Outputter)
			switch item.S.Outputter {
			case "stdout":
				o := new(stdout)
				item.S.Out = o
			}
		}
		item.S.Out.Send(item)
	}
}
