package generator

import (
	"math/rand"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

func Start(gq chan *config.GenQueueItem, gqs chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	gens := make(map[string]config.Generator)
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
		if gens[item.S.Name] == nil {
			log.Infof("Setting sample '%s' to generator '%s'", item.S.Name, item.S.Generator)
			if item.S.Generator == "sample" || item.S.Generator == "replay" {
				s := new(sample)
				gens[item.S.Name] = s
			} else {
				s := new(luagen)
				gens[item.S.Name] = s
			}
			PrimeRater(item.S)
		}
		// log.Debugf("Generating item %#v", item)
		err := gens[item.S.Name].Gen(item)
		if err != nil {
			log.Errorf("Error received from generator: %s", err)
		}
		// log.Debugf("Finished generating item %#v", item)
	}
}
