package main

import (
	"encoding/json"

	"github.com/coccyx/gogen/generator"
	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/timer"
)

func main() {

	// filename := os.Args[1]
	c := config.NewConfig()
	c.Log.Debugf("Global: %#v", c.Global)
	c.Log.Debugf("Default Tokens: %#v", c.DefaultTokens)
	c.Log.Debugf("Default Sample: %#v", c.DefaultSample)
	c.Log.Debugf("Samples: %#v", c.Samples)
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	j, _ := json.MarshalIndent(c, "", "  ")
	c.Log.Debugf("JSON Config: %s\n", j)

	c.Log.Infof("Starting Timers")

	var timers []timer.Timer
	gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)

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

	for i := 0; i < c.Global.OutputWorkers; i++ {
		c.Log.Debugf("Starting Outputter %d", i)
		go outputter.Start(oq)
	}

	for {

	}
}
