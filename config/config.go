package config

import (
	"encoding/json"
	"github.com/coccyx/gogen/timeparser"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
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

			if s.End == "now" || s.End == "" {
				s.realtime = true
			}
			if len(s.Begin) > 0 {
				if s.beginParsed, err = timeparser.TimeParser(s.Begin, time.Now); err != nil {
					c.Log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
					return nil
				}
			}
			if len(s.End) > 0 {
				if s.endParsed, err = timeparser.TimeParser(s.End, time.Now); err != nil {
					c.Log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
					return nil
				}
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
