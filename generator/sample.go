package generator

import "github.com/coccyx/gogen/config"

func SampleGen(item *GenQueueItem) {
	c := config.NewConfig()

	c.Log.Debugf("Gen Queue Item %#v", item)
}
