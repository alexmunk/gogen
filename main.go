package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/ghodss/yaml"
	logging "github.com/op/go-logging"
	"github.com/pkg/profile"
	"gopkg.in/urfave/cli.v1"
)

var c *config.Config

// Setup the running environment
func Setup(clic *cli.Context) {
	if len(clic.String("config")) > 0 {
		os.Setenv("GOGEN_FULLCONFIG", clic.String("config"))
	} else if len(clic.String("samplesDir")) > 0 {
		os.Setenv("GOGEN_SAMPLES_DIR", clic.String("samplesDir"))
	}
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
				cli.StringFlag{Name: "outputter, o"},
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
								c.Samples[i].Output.OutputTemplate = clic.String("outputTemplate")
							}
							if clic.Int("endIntervals") > 0 {
								c.Log.Infof("Setting endIntervals to %d for sample '%s'", clic.Int("endIntervals"), c.Samples[i].Name)
								c.Samples[i].EndIntervals = clic.Int("endIntervals")
							}
							if len(clic.String("outputter")) > 0 {
								c.Log.Infof("Setting outputter to '%s' for sample '%s'", clic.String("outputter"), c.Samples[i].Name)
								c.Samples[i].Output.Outputter = clic.String("outputter")
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
				run.Run(c)
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
		Setup(clic)
		return nil
	}
	app.Action = func(clic *cli.Context) error {
		clic.App.Command("gen").Run(clic)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "info, v",
			Usage: "Sets info level logging",
		},
		cli.BoolFlag{
			Name:  "debug, vv",
			Usage: "Sets debug level logging",
		},
		cli.IntFlag{
			Name:  "generators, g",
			Usage: "Sets number of generator `threads`",
		},
		cli.IntFlag{
			Name:  "outputters, o",
			Usage: "Sets number of outputter `threads`",
		},
		cli.StringFlag{
			Name:  "samplesDir, sd",
			Usage: "Sets `directory` to search for sample files, default 'config/samples'",
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "`Path` or URL to a full config",
		},
	}
	app.Run(os.Args)
}
