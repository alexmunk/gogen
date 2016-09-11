package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/coccyx/timeparser"
	"github.com/ghodss/yaml"
	_ "github.com/hhkbp2/go-strftime"
	"github.com/op/go-logging"
	_ "github.com/pbnjay/strptime"
)

// Config is a struct representing a Singleton which contains a copy of the running config
// across all processes.  Should mirror the structure of $GOGEN_HOME/configs/default/global.yml
type Config struct {
	Global  Global   `json:"global"`
	Samples []Sample `json:"samples"`

	defaultSample Sample `json:"defaultSample"`

	// Exported but internal use variables
	Log      *logging.Logger `json:"-"`
	Timezone *time.Location  `json:"-"`
}

// Global represents global configuration options which apply to all of gogen
type Global struct {
	Debug            bool `json:"debug"`
	Verbose          bool `json:"verbose"`
	UseOutputQueue   bool `json:"useOutputQueue"`
	GeneratorWorkers int  `json:"generatorWorkers"`
	OutputWorkers    int  `json:"outputWorkers"`
}

var instance *Config
var once sync.Once

func getConfig() *Config {
	once.Do(func() {
		instance = &Config{Log: logging.MustGetLogger("gogen")}
	})
	return instance
}

// NewConfig is a singleton constructor which will return a pointer to a global instance of Config
func NewConfig() *Config {
	c := getConfig()
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{color:bold}%{time} %{shortfunc} %{color:%{level:.1s}%{color:reset} %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logging.SetBackend(backendLeveled)

	home := os.Getenv("GOGEN_HOME")
	if len(home) == 0 {
		c.Log.Debug("GOGEN_HOME not set, setting to '.'")
		home = "."
	}
	c.Log.Debugf("Home: %v\n", home)

	// Parse defaults
	if err := parseFileConfig(c, &c.Global, home, "config/default/global.yml"); err != nil {
		c.Log.Panic(err)
	}
	if err := parseFileConfig(c, &c.defaultSample, home, "config/default/sample.yml"); err != nil {
		c.Log.Panic(err)
	}

	// Setup timezone
	c.Timezone, _ = time.LoadLocation("Local")

	// Read all samples in $GOGEN_HOME/config/samples directory
	fullPath := filepath.Join(home, "config", "samples")
	filepath.Walk(fullPath, func(path string, _ os.FileInfo, err error) error {
		innerPath := filepath.Join(fullPath, path)
		c.Log.Debugf("Walking, at %s", innerPath)
		if err != nil {
			c.Log.Errorf("Error from WalkFunc: %s", err)
		}

		// Check if extension is acceptable before attempting to parse
		acceptableExtensions := map[string]int{".yml": 1, ".yaml": 1, ".json": 1}
		if _, ok := acceptableExtensions[filepath.Ext(innerPath)]; ok {
			s := c.defaultSample
			if err := parseFileConfig(c, &s, path); err != nil {
				c.Log.Errorf("Error parsing config %s: %s", innerPath, err)
				return nil
			}

			if len(s.Name) == 0 {
				c.Log.Errorf("Sample from %s is missing name", innerPath)
				return nil
			}

			// Setup Begin & End
			// If End is not set, then we're intended to always run in realtime
			if s.End == "" {
				s.realtime = true
			}
			if len(s.Begin) > 0 {
				if s.beginParsed, err = timeparser.TimeParserNow(s.Begin, time.Now); err != nil {
					c.Log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
				}
			}
			if len(s.End) > 0 {
				if s.endParsed, err = timeparser.TimeParserNow(s.End, time.Now); err != nil {
					c.Log.Errorf("Error parsing End for sample %s: %v", s.Name, err)
				}
			}

			// Parse earliest and latest as relative times

			// Cache a time so we can get a delta for parsed earliest and latest
			n := time.Now()
			now := func() time.Time {
				return n
			}

			var p time.Time
			if p, err = timeparser.TimeParserNow(s.Earliest, now); err != nil {
				c.Log.Errorf("Error parsing earliest time '%s' for sample '%s', using Now", s.Earliest, s.Name)
				s.earliestParsed = time.Duration(0)
			} else {
				s.earliestParsed = n.Sub(p)
			}
			if p, err = timeparser.TimeParserNow(s.Latest, now); err != nil {
				c.Log.Errorf("Error parsing latest time '%s' for sample '%s', using Now", s.Latest, s.Name)
				s.latestParsed = time.Duration(0)
			} else {
				s.latestParsed = n.Sub(p)
			}

			c.Samples = append(c.Samples, s)
		}
		return nil
	})
	return c
}

func parseFileConfig(c *Config, out interface{}, path ...string) error {
	fullPath := filepath.Join(path...)
	c.Log.Debugf("Config Path: %v\n", fullPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return err
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// c.Log.Debugf("Contents: %s", contents)
	switch filepath.Ext(fullPath) {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(contents, out); err != nil {
			c.Log.Panicf("YAML parsing error: %v", err)
		}
	case ".json":
		if err := json.Unmarshal(contents, out); err != nil {
			c.Log.Panicf("JSON parsing error: %v", err)
		}
	}
	// c.Log.Debugf("Out: %#v\n", out)
	return nil
}
