package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/coccyx/gogen/share"
	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	logging "github.com/op/go-logging"
	"github.com/pkg/profile"
	"gopkg.in/urfave/cli.v1"
)

var c *config.Config

// Setup the running environment
func Setup(clic *cli.Context) {
	if len(clic.String("config")) > 0 {
		cstr := clic.String("config")
		if cstr[0:4] == "http" || cstr[len(cstr)-3:] == "yml" || cstr[len(cstr)-4:] == "yaml" || cstr[len(cstr)-4:] == "json" {
			os.Setenv("GOGEN_FULLCONFIG", cstr)
		} else {
			share.PullFile(cstr, ".config.json")
			config.ResetConfig()
			os.Setenv("GOGEN_FULLCONFIG", ".config.json")
			defer func() {
				os.Remove(".config.json")
			}()
		}
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

func table(l []share.GogenList) {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetColWidth(132)
	t.SetAutoWrapText(false)
	t.SetHeader([]string{"Gogen", "Description"})
	for _, li := range l {
		t.Append([]string{li.Gogen, li.Description})
	}
	t.Render()
}

func main() {
	if config.ProfileOn {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
		// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	}
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
				cli.StringFlag{
					Name:  "sample, s",
					Usage: "Only run sample `name`",
				},
				cli.IntFlag{
					Name:  "count, c",
					Usage: "Output `number` events",
				},
				cli.IntFlag{
					Name:  "interval, i",
					Usage: "Output every `seconds` seconds"},
				cli.IntFlag{
					Name:  "endIntervals, ei",
					Usage: "Only run from `number` intervals",
				},
				cli.StringFlag{
					Name:  "outputTemplate, ot",
					Usage: "Use output template `(raw|csv|json|splunkhec)` for formatting output",
				},
				cli.StringFlag{
					Name:  "outputter, o",
					Usage: "Use outputter `(stdout|devnull|file|http) for output",
				},
				cli.StringFlag{
					Name:  "filename, f",
					Usage: "Set `filename`, only usable with file output",
				},
				cli.StringFlag{
					Name:  "url",
					Usage: "Override all endpoint URLs to just `url` url",
				},
				cli.StringFlag{
					Name:  "begin, b",
					Usage: "Set begin time, in relative time syntax (e.g. -60m for minus 60 mins, now for now, etc)",
				},
				cli.StringFlag{
					Name:  "end, e",
					Usage: "Set end time, in relative time syntax (e.g. -60m for minus 60 mins, now for now, etc)",
				},
				cli.BoolFlag{
					Name:  "realtime, r",
					Usage: "Set to real time, don't stop until killed",
				},
			},
			Action: func(clic *cli.Context) error {
				if len(c.Samples) == 0 {
					fmt.Printf("No samples configured, exiting\n")
					os.Exit(1)
				}
				for i := 0; i < len(c.Samples); i++ {
					if len(clic.String("outputTemplate")) > 0 {
						c.Log.Infof("Setting outputTempalte to '%s'", clic.String("outputTemplate"))
						c.Samples[i].Output.OutputTemplate = clic.String("outputTemplate")
					}
					if clic.Int("endIntervals") > 0 {
						c.Log.Infof("Setting endIntervals to %d", clic.Int("endIntervals"))
						c.Samples[i].EndIntervals = clic.Int("endIntervals")
					}
					if len(clic.String("outputter")) > 0 {
						c.Log.Infof("Setting outputter to '%s'", clic.String("outputter"))
						c.Samples[i].Output.Outputter = clic.String("outputter")
					}
					if len(clic.String("filename")) > 0 {
						c.Log.Infof("Setting filename to '%s'")
						c.Samples[i].Output.FileName = clic.String("filename")
					}
					if len(clic.String("url")) > 0 {
						c.Log.Infof("Setting all endpoint urls to '%s'")
						c.Samples[i].Output.Endpoints = []string{clic.String("url")}
					}
					if clic.Int("count") > 0 {
						c.Log.Infof("Setting count to %d for sample '%s'", clic.Int("count"), c.Samples[i].Name)
						c.Samples[i].Count = clic.Int("count")
					}
					if clic.Int("interval") > 0 {
						c.Log.Infof("Setting interval to %d for sample '%s'", clic.Int("interval"), c.Samples[i].Name)
						c.Samples[i].Interval = clic.Int("interval")
					}
					if len(clic.String("begin")) > 0 {
						c.Log.Infof("Setting begin to %s for sample '%s'", clic.String("begin"), c.Samples[i].Name)
						c.Samples[i].Begin = clic.String("begin")
					}
					if len(clic.String("end")) > 0 {
						c.Log.Infof("Setting end to %s for sample '%s'", clic.String("end"), c.Samples[i].Name)
						c.Samples[i].End = clic.String("end")
					}
					if len(clic.String("begin")) > 0 || len(clic.String("end")) > 0 {
						if clic.Int("endIntervals") == 0 {
							c.Samples[i].EndIntervals = 0
						}
						config.ParseBeginEnd(c.Samples[i])
					}
					if clic.Bool("realtime") {
						if clic.Int("endIntervals") == 0 {
							c.Samples[i].EndIntervals = 0
						}
						c.Samples[i].Realtime = true
					}
				}
				if len(clic.String("sample")) > 0 {
					c.Log.Infof("Generating only for sample '%s'", clic.String("sample"))
					matched := false
					for i := 0; i < len(c.Samples); i++ {
						if c.Samples[i].Name == clic.String("sample") {
							matched = true
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
				os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
				os.Setenv("GOGEN_EXPORT", "1")
				c = config.NewConfig()
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
		{
			Name:  "login",
			Usage: "Login to GitHub",
			Action: func(clic *cli.Context) error {
				_ = share.NewGitHub(true)
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List all published Gogens",
			Action: func(clic *cli.Context) error {
				fmt.Printf("Showing all Gogens:\n\n")
				l := share.List()
				table(l)
				return nil
			},
		},
		{
			Name:  "search",
			Usage: "Search for Gogens",
			Action: func(clic *cli.Context) error {
				var q string
				for _, a := range clic.Args() {
					q += a + " "
				}
				q = strings.TrimRight(q, " ")
				fmt.Printf("Returning results for search: \"%s\"\n\n", q)
				l := share.Search(q)
				if len(l) > 0 {
					table(l)
				} else {
					fmt.Println("   No results found.")
				}
				return nil
			},
		},
		{
			Name:      "info",
			Usage:     "Get info on a specific Gogen",
			ArgsUsage: "[owner/name]",
			Action: func(clic *cli.Context) error {
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a Gogen in owner/name format")
					os.Exit(1)
				}
				g := share.Get(clic.Args()[0])
				fmt.Printf("Details for Gogen %s\n", g.Gogen)
				fmt.Printf("------------------------------------------------------\n")
				fmt.Printf("%15s : %s\n", "Gogen", g.Gogen)
				fmt.Printf("%15s : %s\n", "Owner", g.Owner)
				fmt.Printf("%15s : %s\n", "Name", g.Name)
				fmt.Printf("%15s : %s\n\n", "Description", g.Description)
				if len(g.Notes) > 0 {
					fmt.Printf("Notes:\n")
					fmt.Printf("------------------------------------------------------\n")
					fmt.Printf("%s\n", g.Notes)
				}
				var event map[string]interface{}
				var eventbytes []byte
				_ = json.Unmarshal([]byte(g.SampleEvent), &event)
				eventbytes, _ = json.MarshalIndent(event, "", "  ")
				fmt.Printf("Sample Event:\n")
				fmt.Printf("------------------------------------------------------\n")
				fmt.Printf("%s\n", string(eventbytes))
				return nil
			},
		},
		{
			Name:  "push",
			Usage: "Push running config to Gogen sharing service",
			ArgsUsage: "[name]\n\n" + `This will push your running config to the Gogen sharing API.  This will publish the running config in a Git Gist and make an entry in the
Gogen API database pointing to the gist with a bit of metadata.app

The [name] argument should be the name of the primary sample you are publishing.  The entry in the database will get its Name, Description and Notes
from the sample referenced by [name]`,
			Action: func(clic *cli.Context) error {
				config.ResetConfig()
				os.Setenv("GOGEN_EXPORT", "1")
				_ = config.NewConfig()
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a name to publish this config")
					os.Exit(1)
				}
				owner, id := share.Push(clic.Args().First())
				fmt.Printf("Push successful.  Gist: https://gist.github.com/%s/%s\n", owner, id)
				return nil
			},
		},
		{
			Name:      "pull",
			Usage:     "Pull a config down for editing",
			ArgsUsage: "[owner/name] [directory]",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "deconstruct, d"},
			},
			Action: func(clic *cli.Context) error {
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a Gogen in owner/name format")
					os.Exit(1)
				} else if len(clic.Args()) < 2 {
					fmt.Println("Error: Must specify a directory to place config files")
					os.Exit(1)
				}
				share.Pull(clic.Args()[0], clic.Args()[1], clic.Bool("deconstruct"))
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
