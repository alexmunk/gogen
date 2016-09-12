package main

import (
	"github.com/coccyx/gogen/config"
	"github.com/coccyx/gogen/generator"
	"github.com/coccyx/gogen/timer"
)

func main() {

	// filename := os.Args[1]
	c := config.NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	// j, _ := json.MarshalIndent(c, "", "  ")
	// c.Log.Debugf("JSON Config: %s\n", j)

	c.Log.Infof("Starting Timers")

	var timers []timer.Timer
	gq := make(chan *generator.GenQueueItem)
	oq := make(chan []string)

	for i := 0; i < len(c.Samples); i++ {
		s := c.Samples[i]
		if !s.Disabled {
			t := timer.Timer{S: s, GQ: gq, OQ: oq}
			go t.NewTimer()
			timers = append(timers, t)
		}
	}

	c.Log.Infof("Starting Generators")

	for i := 0; i < c.Global.GeneratorWorkers; i++ {
		c.Log.Debugf("Starting Generator %d", i)
		go generator.Start(gq)
	}

	for {

	}
}
