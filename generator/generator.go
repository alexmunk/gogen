package generator

import "github.com/coccyx/gogen/internal"

func Start(gq chan *config.GenQueueItem) {
	for {
		item := <-gq
		// Check to see if our generator is not set
		if item.S.Gen == nil {
			item.S.Log.Infof("Setting sample '%s' to generator '%s'", item.S.Name, item.S.Generator)
			if item.S.Generator == "sample" {
				s := new(sample)
				item.S.Gen = s
			}
		}
		item.S.Gen.Gen(item)
	}
}
