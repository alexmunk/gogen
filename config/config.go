package config

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Config is a struct representing a Singleton which contains a copy of the running config
// across all processes.  Should mirror the structure of $GOGEN_HOME/configs/default/global.yml
type Config struct {
	Global        Global
	defaultSample Sample
	samples       []Sample
	Log           *logging.Logger
}

type Global struct {
	Debug            bool `json:"debug"`
	Verbose          bool `json:"verbose"`
	UseOutputQueue   bool `json:"useOutputQueue"`
	GeneratorWorkers int  `json:"generatorWorkers"`
	OutputWorkers    int  `json:"outputWorkers"`
}

type Sample struct {
	Disabled        bool    `json:"disabled"`
	Generator       string  `json:"generator"`
	Outputter       string  `json:"outputter"`
	Rater           string  `json:"rater"`
	Interval        int     `json:"interval"`
	Delay           int     `json:"delay"`
	Count           int     `json:"count"`
	Earliest        string  `json:"earliest"`
	Latest          string  `json:"latest"`
	RandomizeCount  float32 `json:"randomizeCount"`
	RandomizeEvents bool    `json:"randomizeEvents"`
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
	var _ = json.Unmarshal // TODO: Unused, delete when done

	home := os.Getenv("GOGEN_HOME")
	if len(home) == 0 {
		c.Log.Debug("GOGEN_HOME not set, setting to '.'")
		home = "."
	}
	c.Log.Debugf("Home: %v\n", home)

	// Parse defaults
	if err := parseConfig(c, &c.Global, home, "config/default/global.yml"); err != nil {
		c.Log.Panic(err)
	}
	if err := parseConfig(c, &c.defaultSample, home, "config/default/sample.yml"); err != nil {
		c.Log.Panic(err)
	}

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
			if err := parseConfig(c, &s, path); err != nil {
				c.Log.Errorf("Error parsing config %s: %s", innerPath, err)
				return nil
			}
			c.samples = append(c.samples, s)
		}
		return nil
	})
	return c
}

func parseConfig(c *Config, out interface{}, path ...string) error {
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
			c.Log.Panicf("Yaml parsing error: %v", err)
		}
	case ".json":
		if err := json.Unmarshal(contents, out); err != nil {
			c.Log.Panicf("JSON parsing error: %v", err)
		}
	}
	// c.Log.Debugf("Out: %#v\n", out)
	return nil
}
