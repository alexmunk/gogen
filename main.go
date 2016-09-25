package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/coccyx/gogen/generator"
	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/timer"
	"github.com/ghodss/yaml"
	logging "github.com/op/go-logging"
	"github.com/pkg/profile"
	"gopkg.in/urfave/cli.v1"
)

var c *config.Config

func setup(clic *cli.Context) {
	c = config.NewConfig()

	if clic.Bool("debug") {
		c.SetLoggingLevel(logging.DEBUG)
	} else if clic.Bool("info") {
		c.SetLoggingLevel(logging.INFO)
	}

	if clic.Int("generators") > 0 {
		c.Log.Infof("Setting generators to %d", clic.Int("generators"))
		c.Global.GeneratorWorkers = clic.Int("generators")
	}
	if clic.Int("outputters") > 0 {
		c.Log.Infof("Setting generators to %d", clic.Int("outputters"))
		c.Global.OutputWorkers = clic.Int("outputters")
	}

	// c.Log.Debugf("Global: %#v", c.Global)
	// c.Log.Debugf("Default Tokens: %#v", c.DefaultTokens)
	// c.Log.Debugf("Default Sample: %#v", c.DefaultSample)
	// c.Log.Debugf("Samples: %#v", c.Samples)
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	// j, _ := json.MarshalIndent(c, "", "  ")
	// c.Log.Debugf("JSON Config: %s\n", j)
}

func gen() {
	c.Log.Info("Starting Timers")
	timerdone := make(chan int)
	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)
	gens := 0
	outs := 0
	timers := 0
	for i := 0; i < len(c.Samples); i++ {
		s := c.Samples[i]
		if !s.Disabled {
			t := timer.Timer{S: s, GQ: gq, OQ: oq, Done: timerdone}
			go t.NewTimer()
			timers++
		}
	}
	c.Log.Infof("%d Timers started", timers)

	c.Log.Infof("Starting Generators")
	for i := 0; i < c.Global.GeneratorWorkers; i++ {
		c.Log.Infof("Starting Generator %d", i)
		go generator.Start(gq, gqs)
		gens++
	}

	c.Log.Infof("Starting Outputters")
	for i := 0; i < c.Global.OutputWorkers; i++ {
		c.Log.Infof("Starting Outputter %d", i)
		go outputter.Start(oq, oqs)
		outs++
	}

	// time.Sleep(1000 * time.Millisecond)

	// Check if any timers are done
Loop1:
	for {
		select {
		case <-timerdone:
			timers--
			c.Log.Debugf("Timer done, timers left %d", timers)
			if timers == 0 {
				break Loop1
			}
		}
	}

	// Close our channels to signal to the workers to shut down when the queue is clear
	c.Log.Infof("Timers all done, closing generating queue")
	close(gq)

	// Check for all the workers to signal back they're done
Loop2:
	for {
		select {
		case <-gqs:
			gens--
			c.Log.Debugf("Gen done, gens left %d", gens)
			if gens == 0 {
				break Loop2
			}
		}
	}

	// Close our output channel to signal to outputters we're done
	close(oq)
Loop3:
	for {
		select {
		case <-oqs:
			outs--
			c.Log.Debugf("Out done, outs left %d", outs)
			if outs == 0 {
				break Loop3
			}
		}
	}

	// time.Sleep(100 * time.Millisecond)
}

func main() {
	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	rand.Seed(time.Now().UnixNano())

	app := cli.NewApp()
	app.Name = "gogen"
	app.Usage = "Generate data for demos and testing"
	app.Version = "0.1.0"
	cli.VersionFlag = cli.BoolFlag{Name: "version"}
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Clint Sharp",
			Email: "clint@typhoon.org",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "gen",
			Usage: "Generate Events",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "sample, s"},
				cli.IntFlag{Name: "count, c"},
				cli.IntFlag{Name: "interval, i"},
				cli.IntFlag{Name: "endIntervals, ei"},
				cli.StringFlag{Name: "outputTemplate, ot"},
			},
			Action: func(clic *cli.Context) error {
				if len(clic.String("sample")) > 0 {
					c.Log.Infof("Generating only for sample '%s'", clic.String("sample"))
					matched := false
					for i := 0; i < len(c.Samples); i++ {
						if c.Samples[i].Name == clic.String("sample") {
							matched = true
							if clic.Int("count") > 0 {
								c.Log.Infof("Setting count to %d for sample '%s'", clic.Int("count"), c.Samples[i].Name)
								c.Samples[i].Count = clic.Int("count")
							}
							if clic.Int("interval") > 0 {
								c.Log.Infof("Setting interval to %d for sample '%s'", clic.Int("interval"), c.Samples[i].Name)
								c.Samples[i].Interval = clic.Int("interval")
							}
							if len(clic.String("outputTemplate")) > 0 {
								c.Log.Infof("Setting outputTempalte to '%s' for sample '%s'", clic.String("outputTemplate"), c.Samples[i].Name)
								c.Samples[i].OutputTemplate = clic.String("outputTemplate")
							}
							if clic.Int("endIntervals") > 0 {
								c.Log.Infof("Setting endIntervals to %d for sample '%s'", clic.Int("endIntervals"), c.Samples[i].Name)
								c.Samples[i].EndIntervals = clic.Int("endIntervals")
							}
						} else {
							c.Samples[i].Disabled = true
						}
					}
					if !matched {
						c.Log.Errorf("No sample matched for '%s'", clic.String("sample"))
						os.Exit(1)
					}
				}
				gen()
				return nil
			},
		},
		{
			Name:  "config",
			Usage: "Print config to stdout",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format, f"},
			},
			Action: func(clic *cli.Context) error {
				var outb []byte
				var err error
				if clic.String("format") == "yaml" {
					if outb, err = yaml.Marshal(c); err != nil {
						c.Log.Panicf("YAML output error: %v", err)
					}
				} else {
					if outb, err = json.MarshalIndent(c, "", "  "); err != nil {
						c.Log.Panicf("JSON output error: %v", err)
					}
				}
				out := string(outb)
				fmt.Print(out)
				return nil
			},
		},
	}
	app.Before = func(clic *cli.Context) error {
		setup(clic)
		return nil
	}
	app.Action = func(clic *cli.Context) error {
		clic.App.Command("gen").Run(clic)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "info, v"},
		cli.BoolFlag{Name: "debug, vv"},
		cli.IntFlag{Name: "generators, g"},
		cli.IntFlag{Name: "outputters, o"},
	}
	app.Run(os.Args)
}
