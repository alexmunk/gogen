package generator

import (
	"math/rand"
	"time"

	"github.com/coccyx/gogen/internal"
)

func Start(gq chan *config.GenQueueItem, gqs chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	for {
		item, ok := <-gq
		if !ok {
			gqs <- 1
			break
		}
		item.Rand = generator
		// Check to see if our generator is not set
		if item.S.Gen == nil {
			item.S.Log.Infof("Setting sample '%s' to generator '%s'", item.S.Name, item.S.Generator)
			if item.S.Generator == "sample" {
				s := new(sample)
				item.S.Gen = s
			}
		}
		item.S.Log.Debugf("Generating item %#v", item)
		item.S.Gen.Gen(item)
		item.S.Log.Debugf("Finished generating item %#v", item)
	}
}
