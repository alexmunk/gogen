package generator

import (
	"math/rand"
	"time"

	"github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
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
			log.Infof("Setting sample '%s' to generator '%s'", item.S.Name, item.S.Generator)
			if item.S.Generator == "sample" {
				s := new(sample)
				item.S.Gen = s
			}
		}
		log.Debugf("Generating item %#v", item)
		item.S.Gen.Gen(item)
		log.Debugf("Finished generating item %#v", item)
	}
}
