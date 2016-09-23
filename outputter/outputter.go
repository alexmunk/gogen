package outputter

import (
	"github.com/coccyx/gogen/internal"
)

func Start(oq chan *config.OutQueueItem) {
	for {
		item := <-oq
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
