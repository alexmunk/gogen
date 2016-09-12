package config

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
	Global        Global    `json:"global"`
	Samples       []*Sample `json:"samples"`
	DefaultTokens []*Token  `json:"defaultTokens"`

	defaultSample Sample
	initialized   bool

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
		instance = &Config{Log: logging.MustGetLogger("gogen"), initialized: false}
	})
	return instance
}

// NewConfig is a singleton constructor which will return a pointer to a global instance of Config
func NewConfig() *Config {
	c := getConfig()
	if c.initialized {
		return c
	}
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
	if err := c.parseFileConfig(&c.Global, home, "config/default/global.yml"); err != nil {
		c.Log.Panic(err)
	}
	if err := c.parseFileConfig(&c.defaultSample, home, "config/default/sample.yml"); err != nil {
		c.Log.Panic(err)
	}

	// Setup timezone
	c.Timezone, _ = time.LoadLocation("Local")

	// Read all default tokens in $GOGEN_HOME/config/default/tokens
	fullPath := filepath.Join(home, "config", "default", "tokens")
	acceptableExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true}
	c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
		t := new(Token)

		if err := c.parseFileConfig(&t, innerPath); err != nil {
			c.Log.Errorf("Error parsing config %s: %s", innerPath, err)
			return err
		}

		c.DefaultTokens = append(c.DefaultTokens, t)
		return nil
	})

	// Read all flat file samples
	fullPath = filepath.Join(home, "config", "samples")
	acceptableExtensions = map[string]bool{".sample": true}
	c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
		s := new(Sample)
		s.Name = filepath.Base(innerPath)
		s.Disabled = true

		file, err := os.Open(innerPath)
		if err != nil {
			c.Log.Errorf("Error reading sample file '%s': %s", innerPath, err)
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s.Lines = append(s.Lines, scanner.Text())
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all csv file samples
	fullPath = filepath.Join(home, "config", "samples")
	acceptableExtensions = map[string]bool{".csv": true}
	c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
		s := new(Sample)
		s.Name = filepath.Base(innerPath)
		s.Disabled = true

		var (
			fields []string
			rows   [][]string
			err    error
		)

		file, err := os.Open(innerPath)
		if err != nil {
			c.Log.Errorf("Error reading sample file '%s': %s", innerPath, err)
			return nil
		}
		defer file.Close()

		reader := csv.NewReader(file)
		if fields, err = reader.Read(); err != nil {
			c.Log.Errorf("Error parsing header row of sample file '%s' as csv: %s", innerPath, err)
			return nil
		}
		if rows, err = reader.ReadAll(); err != nil {
			c.Log.Errorf("Error parsing sample file '%s' as csv: %s", innerPath, err)
			return nil
		}
		for _, row := range rows {
			fieldsmap := map[string]string{}
			for i := 0; i < len(fields); i++ {
				if fields[i] == "_raw" {
					s.Lines = append(s.Lines, row[i])
				} else {
					fieldsmap[fields[i]] = row[i]
				}
			}
			s.LinesMap = append(s.LinesMap, fieldsmap)
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all YAML & JSON samples in $GOGEN_HOME/config/samples directory
	fullPath = filepath.Join(home, "config", "samples")
	acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
	c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
		s := c.defaultSample
		if err := c.parseFileConfig(&s, innerPath); err != nil {
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
			s.Realtime = true
		}
		var err error
		if len(s.Begin) > 0 {
			if s.BeginParsed, err = timeparser.TimeParserNow(s.Begin, time.Now); err != nil {
				c.Log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
			}
		}
		if len(s.End) > 0 {
			if s.EndParsed, err = timeparser.TimeParserNow(s.End, time.Now); err != nil {
				c.Log.Errorf("Error parsing End for sample %s: %v", s.Name, err)
			}
		}

		//
		// Parse earliest and latest as relative times
		//

		// Cache a time so we can get a delta for parsed earliest and latest
		n := time.Now()
		now := func() time.Time {
			return n
		}

		var p time.Time
		if p, err = timeparser.TimeParserNow(s.Earliest, now); err != nil {
			c.Log.Errorf("Error parsing earliest time '%s' for sample '%s', using Now", s.Earliest, s.Name)
			s.EarliestParsed = time.Duration(0)
		} else {
			s.EarliestParsed = n.Sub(p) * -1
		}
		if p, err = timeparser.TimeParserNow(s.Latest, now); err != nil {
			c.Log.Errorf("Error parsing latest time '%s' for sample '%s', using Now", s.Latest, s.Name)
			s.LatestParsed = time.Duration(0)
		} else {
			s.LatestParsed = n.Sub(p) * -1
		}

		//
		// Setup tokens from defaults
		//

		// Iterate through all tokens, then for each token we will scan all default tokens for a match
		for i := 0; i < len(s.Tokens); i++ {
			t := &s.Tokens[i]
			tf := reflect.ValueOf(t).Elem()
			// typeOfT := tf.Type()

			// Iterate through all DefaultTokens looking for a name match
			for _, dt := range c.DefaultTokens {
				// Name matches
				if dt.Name == t.Name {
					// c.Log.Debugf("Token names match: %s %s\n", t.Name, dt.Name)
					// c.Log.Debugf("Value of source %#v", t)
					// Iterate through the token's fields
					for fi := 0; fi < tf.NumField(); fi++ {
						// c.Log.Debugf("Comparing field %s\n", typeOfT.Field(fi).Name)
						f := tf.Field(fi)                               // Set current field of actual token
						sourcef := reflect.ValueOf(dt).Elem().Field(fi) // Set field value to copy from, if we can
						// Override value if value is blank
						// Check for blankness based on type
						switch f.Kind() {
						case reflect.Int:
							// c.Log.Debugf("Field '%s' value '%d'", typeOfT.Field(fi).Name, f.Int())
							if f.Int() == 0 {
								// c.Log.Debugf("Setting source to %d", sourcef.Int())
								f.SetInt(sourcef.Int())
							}
						case reflect.String:
							// c.Log.Debugf("Field '%s' value '%s'", typeOfT.Field(fi).Name, f.String())
							if f.String() == "" {
								// c.Log.Debugf("Setting source to %s", sourcef.String())
								f.SetString(sourcef.String())
							}
						case reflect.Map:
							// c.Log.Debugf("Field '%s' is map", typeOfT.Field(fi).Name)
							if f.Len() == 0 {
								// c.Log.Debugf("Setting map for field '%s' for token '%s'", typeOfT.Field(fi).Name, t.Name)
								// If it is a map we create a new map and translate each value
								f.Set(reflect.MakeMap(sourcef.Type()))
								for _, key := range sourcef.MapKeys() {
									sourceValue := sourcef.MapIndex(key)
									// New gives us a pointer, but again we want the value
									destValue := reflect.New(sourceValue.Type()).Elem()
									f.SetMapIndex(key, destValue)
								}
							}
						case reflect.Array:
							// c.Log.Debugf("Field '%s' is array", typeOfT.Field(fi).Name)
							if f.Len() == 0 {
								reflect.Copy(f, sourcef)
							}
						}
					}
					// c.Log.Debugf("New Token value: %#v", t)
					// c.Log.Debugf("New tokens values: %#v", s.Tokens)
				}
			}

		}

		c.Samples = append(c.Samples, &s)
		return nil
	})

	// There area references from tokens to samples, need to resolve those references
	for i := 0; i < len(c.Samples); i++ {
		c.resolve(c.Samples[i])
	}

	c.initialized = true
	return c
}

// resolve takes a sample, finds any references from tokens to other samples and
// updates the token to point to the sample data
func (c *Config) resolve(s *Sample) {
	// c.Log.Debugf("Resolving '%s'", s.Name)
	for i := 0; i < len(s.Tokens); i++ {
		// c.Log.Debugf("Resolving token '%s' for sample '%s'", s.Tokens[i].Name, s.Name)
		for j := 0; j < len(c.Samples); j++ {
			if s.Tokens[i].Sample == c.Samples[j].Name {
				c.Log.Debugf("Resolving sample '%s' to token '%s'", c.Samples[j].Name, s.Tokens[i].Sample)
				s.Tokens[i].FieldChoice = c.Samples[j].LinesMap
				s.Tokens[i].PercChoice = c.Samples[j].LinesMap
				s.Tokens[i].Choice = c.Samples[j].Lines
				break
			}
		}
	}
}

func (c *Config) walkPath(fullPath string, acceptableExtensions map[string]bool, callback func(string) error) error {
	filepath.Walk(fullPath, func(path string, _ os.FileInfo, err error) error {
		c.Log.Debugf("Walking, at %s", path)
		if err != nil {
			c.Log.Errorf("Error from WalkFunc: %s", err)
			return err
		}
		// Check if extension is acceptable before attempting to parse
		if acceptableExtensions[filepath.Ext(path)] {
			return callback(path)
		}
		return nil
	})
	return nil
}

func (c *Config) parseFileConfig(out interface{}, path ...string) error {
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
