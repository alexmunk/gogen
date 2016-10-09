package config

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/coccyx/gogen/template"
	"github.com/coccyx/timeparser"
	"github.com/ghodss/yaml"
	"github.com/op/go-logging"
)

// Config is a struct representing a Singleton which contains a copy of the running config
// across all processes.  Should mirror the structure of $GOGEN_HOME/configs/default/global.yml
type Config struct {
	Global      Global      `json:"global"`
	Samples     []*Sample   `json:"samples"`
	Templates   []*Template `json:"templates"`
	initialized bool

	// Exported but internal use variables
	Log      *logging.Logger `json:"-"`
	Timezone *time.Location  `json:"-"`
	Buf      bytes.Buffer    `json:"-"`
}

// Global represents global configuration options which apply to all of gogen
type Global struct {
	Debug            bool   `json:"debug"`
	Verbose          bool   `json:"verbose"`
	GeneratorWorkers int    `json:"generatorWorkers"`
	OutputWorkers    int    `json:"outputWorkers"`
	ROTInterval      int    `json:"rotInterval"`
	Output           Output `json:"output"`
	SamplesDir       []string
}

// Output represents configuration for outputting data
type Output struct {
	FileName       string            `json:"fileName"`
	MaxBytes       int64             `json:"maxBytes"`
	BackupFiles    int               `json:"backupFiles"`
	BufferBytes    int               `json:"bufferBytes"`
	Outputter      string            `json:"outputter"`
	OutputTemplate string            `json:"outputTemplate"`
	Endpoints      []string          `json:"endpoints"`
	Headers        map[string]string `json:"headers"`
}

var instance *Config
var once sync.Once

func getConfig() *Config {
	// once.Do(func() {
	// 	instance = &Config{Log: logging.MustGetLogger("gogen"), initialized: false}
	// })
	if instance == nil {
		instance = &Config{Log: logging.MustGetLogger("gogen"), initialized: false}
	}
	return instance
}

// ResetConfig will delete any current running config
func ResetConfig() {
	instance = nil
}

// NewConfig is a singleton constructor which will return a pointer to a global instance of Config
// Environment variables will impact the function of how we configure ourselves
// GOGEN_HOME: Change home directory where we will search for configs
// GOGEN_SAMPLES_DIR: Change where we will look for samples
// GOGEN_ALWAYS_REFRESH: Do not use singleton pattern and reparse configs
// GOGEN_FULLCONFIG: The reference is to a full exported config, so don't resolve or validate
// GOGEN_EXPORT: Don't set defaults for export
func NewConfig() *Config {
	var c *Config
	if os.Getenv("GOGEN_ALWAYS_REFRESH") != "1" {
		c = getConfig()
		if c.initialized {
			return c
		}
	} else {
		c = &Config{Log: logging.MustGetLogger("gogen"), initialized: false}
		c.Log.Debugf("Always refresh on, using fresh config")
	}

	c.SetLoggingLevel(DefaultLoggingLevel)
	// Setup timezone
	c.Timezone, _ = time.LoadLocation("Local")

	home := os.Getenv("GOGEN_HOME")
	if len(home) == 0 {
		c.Log.Debug("GOGEN_HOME not set, setting to '.'")
		home = "."
		os.Setenv("GOGEN_HOME", home)
	}
	c.Log.Debugf("Home: %v\n", home)

	samplesDir := os.Getenv("GOGEN_SAMPLES_DIR")
	if len(samplesDir) == 0 {
		samplesDir = filepath.Join(home, "config", "samples")
		c.Log.Debugf("GOGEN_SAMPLES_DIR not set, setting to '%s'", samplesDir)
	}

	fullConfig := os.Getenv("GOGEN_FULLCONFIG")
	if len(fullConfig) > 0 {
		if fullConfig[0:4] == "http" {
			c.Log.Infof("Fetching config from '%s'", fullConfig)
			if err := c.parseWebConfig(&c, fullConfig); err != nil {
				c.Log.Panic(err)
			}
		} else {
			_, err := os.Stat(fullConfig)
			if err != nil {
				c.Log.Fatalf("Cannot stat file %s", fullConfig)
			}
			if err := c.parseFileConfig(&c, fullConfig); err != nil {
				c.Log.Panic(err)
			}
			c.Global.SamplesDir = append(c.Global.SamplesDir, filepath.Dir(fullConfig))
		}
		for i := 0; i < len(c.Samples); i++ {
			c.Samples[i].realSample = true
		}
	} else {
		globalFile := os.Getenv("GOGEN_GLOBAL")
		if len(globalFile) > 0 {
			if err := c.parseFileConfig(&c.Global, globalFile); err != nil {
				c.Log.Panic(err)
			}
		}
	}
	if c.Global.Debug {
		c.SetLoggingLevel(logging.DEBUG)
	}

	// Don't set defaults if we're exporting
	if os.Getenv("GOGEN_EXPORT") != "1" {
		//
		// Setup defaults for global
		//
		if c.Global.GeneratorWorkers == 0 {
			c.Global.GeneratorWorkers = defaultGeneratorWorkers
		}
		if c.Global.OutputWorkers == 0 {
			c.Global.OutputWorkers = defaultOutputWorkers
		}
		if c.Global.ROTInterval == 0 {
			c.Global.ROTInterval = defaultROTInterval
		}
		if c.Global.Output.Outputter == "" {
			c.Global.Output.Outputter = defaultOutputter
		}
		if c.Global.Output.OutputTemplate == "" {
			c.Global.Output.OutputTemplate = defaultOutputTemplate
		}

		//
		// Setup defaults for outputs
		//
		switch c.Global.Output.Outputter {
		case "file":
			if c.Global.Output.FileName == "" {
				c.Global.Output.FileName = defaultFileName
			}
			if c.Global.Output.BackupFiles == 0 {
				c.Global.Output.BackupFiles = defaultBackupFiles
			}
			if c.Global.Output.MaxBytes == 0 {
				c.Global.Output.MaxBytes = defaultMaxBytes
			}
		case "http":
			if c.Global.Output.BufferBytes == 0 {
				c.Global.Output.BufferBytes = defaultBufferBytes
			}
		}

		// Add default templates
		templates := []*Template{defaultCSVTemplate, defaultJSONTemplate, defaultSplunkHECTemplate, defaultRawTemplate}
		for _, t := range templates {
			if len(t.Header) > 0 {
				_ = template.New(t.Name+"_header", t.Header)
			}
			_ = template.New(t.Name+"_row", t.Row)
			if len(t.Footer) > 0 {
				_ = template.New(t.Name+"_footer", t.Footer)
			}

			c.Templates = append(c.Templates, t)
		}
	}

	if len(fullConfig) == 0 {
		// Read all templates in $GOGEN_HOME/config/templates
		fullPath := filepath.Join(home, "config", "templates")
		acceptableExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true}
		c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
			t := new(Template)

			if err := c.parseFileConfig(&t, innerPath); err != nil {
				c.Log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			_ = template.New(t.Name+"_header", t.Row)
			_ = template.New(t.Name+"_row", t.Row)
			_ = template.New(t.Name+"_footer", t.Footer)

			c.Templates = append(c.Templates, t)
			return nil
		})

		c.readSamplesDir(samplesDir)
	}

	// Configuration allows for finding additional samples directories and reading them
	for _, sd := range c.Global.SamplesDir {
		c.Log.Debugf("Reading samplesDir: %s", sd)
		c.readSamplesDir(sd)
	}

	// Add a clause to allow copying from other samples
	for i := 0; i < len(c.Samples); i++ {
		if len(c.Samples[i].FromSample) > 0 {
			for j := 0; j < len(c.Samples); j++ {
				if c.Samples[j].Name == c.Samples[i].FromSample {
					c.Log.Debugf("Copying sample '%s' to sample '%s' because fromSample set", c.Samples[j].Name, c.Samples[i].Name)
					tempname := c.Samples[i].Name
					tempcount := c.Samples[i].Count
					tempinterval := c.Samples[i].Interval
					tempendintervals := c.Samples[i].EndIntervals
					tempbegin := c.Samples[i].Begin
					tempend := c.Samples[i].End
					temp := *c.Samples[j]
					c.Samples[i] = &temp
					c.Samples[i].Disabled = false
					c.Samples[i].Name = tempname
					c.Samples[i].FromSample = ""
					if tempcount > 0 {
						c.Samples[i].Count = tempcount
					}
					if tempinterval > 0 {
						c.Samples[i].Interval = tempinterval
					}
					if tempendintervals > 0 {
						c.Samples[i].EndIntervals = tempendintervals
					}
					if len(tempbegin) > 0 {
						c.Samples[i].Begin = tempbegin
					}
					if len(tempend) > 0 {
						c.Samples[i].End = tempend
					}
					break
				}
			}
		}
	}

	// There area references from tokens to samples, need to resolve those references
	for i := 0; i < len(c.Samples); i++ {
		c.validate(c.Samples[i])
	}

	// Clean up disabled and informational samples
	samples := make([]*Sample, 0, len(c.Samples))
	for i := 0; i < len(c.Samples); i++ {
		if c.Samples[i].realSample && !c.Samples[i].Disabled {
			samples = append(samples, c.Samples[i])
		}
	}
	c.Samples = samples

	c.initialized = true
	return c
}

// SetLoggingLevel sets the logging level for everyone
func (c *Config) SetLoggingLevel(level logging.Level) {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(
		`%{color:bold}%{time} %{shortfunc} %{color:%{level:.1s}%{color:reset} %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(level, "")

	logging.SetBackend(backendLeveled)
}

func (c *Config) readSamplesDir(samplesDir string) {
	// Read all flat file samples
	acceptableExtensions := map[string]bool{".sample": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
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
			s.Lines = append(s.Lines, map[string]string{"_raw": scanner.Text()})
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all csv file samples
	acceptableExtensions = map[string]bool{".csv": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
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
				fieldsmap[fields[i]] = row[i]
			}
			s.Lines = append(s.Lines, fieldsmap)
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all YAML & JSON samples in $GOGEN_HOME/config/samples directory
	acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
		s := Sample{}
		if err := c.parseFileConfig(&s, innerPath); err != nil {
			c.Log.Errorf("Error parsing config %s: %s", innerPath, err)
			return nil
		}
		s.realSample = true

		c.Samples = append(c.Samples, &s)
		return nil
	})
}

// validate takes a sample and checks against any rules which may cause the configuration to be invalid.
// This hopefully centralizes logic for valid configs, disabling any samples which are not valid and
// preventing this logic from sprawling all over the code base.
// Also finds any references from tokens to other samples and
// updates the token to point to the sample data
// Also fixes up any additional things which are needed, like weighted choice string
// string map to the randutil Choice struct
func (c *Config) validate(s *Sample) {
	if s.realSample {
		if len(s.Name) == 0 {
			s.Disabled = true
			s.realSample = false
		} else {
			s.realSample = true
		}

		// Give us a logger we can use elsewhere
		s.Log = c.Log

		// Put the output into the sample for convenience
		s.Output = &c.Global.Output

		// Setup defaults
		if s.Generator == "" {
			s.Generator = defaultGenerator
		}
		if s.Earliest == "" {
			s.Earliest = defaultEarliest
		}
		if s.Latest == "" {
			s.Latest = defaultLatest
		}
		if s.RandomizeEvents == false {
			s.RandomizeEvents = defaultRandomizeEvents
		}
		if s.Field == "" {
			s.Field = defaultField
		}

		// Setup Begin & End
		// If End is not set, then we're intended to always run in realtime
		if s.Begin == "" && s.End == "" {
			s.Realtime = true
		}
		// Cache a time so we can get a delta for parsed begin, end, earliest and latest
		n := time.Now()
		now := func() time.Time {
			return n
		}
		var err error
		if len(s.Begin) > 0 {
			if s.BeginParsed, err = timeparser.TimeParserNow(s.Begin, now); err != nil {
				c.Log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
			} else {
				s.Current = s.BeginParsed
				s.Realtime = false
			}
		}
		if len(s.End) > 0 {
			if s.EndParsed, err = timeparser.TimeParserNow(s.End, now); err != nil {
				c.Log.Errorf("Error parsing End for sample %s: %v", s.Name, err)
			}
		}

		//
		// Parse earliest and latest as relative times
		//

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

		// c.Log.Debugf("Resolving '%s'", s.Name)
		for i := 0; i < len(s.Tokens); i++ {
			if s.Tokens[i].Field == "" {
				s.Tokens[i].Field = s.Field
			}
			// If format is template, then create a default token of $tokenname$
			if s.Tokens[i].Format == "template" && s.Tokens[i].Token == "" {
				s.Tokens[i].Token = "$" + s.Tokens[i].Name + "$"
			}
			// c.Log.Debugf("Resolving token '%s' for sample '%s'", s.Tokens[i].Name, s.Name)
			for j := 0; j < len(c.Samples); j++ {
				if s.Tokens[i].SampleString == c.Samples[j].Name {
					c.Log.Debugf("Resolving sample '%s' for token '%s'", c.Samples[j].Name, s.Tokens[i].Name)
					s.Tokens[i].Sample = c.Samples[j]
					// See if a field exists other than _raw, if so, FieldChoice
					otherfield := false
					if len(c.Samples[j].Lines) > 0 {
						for k := range c.Samples[j].Lines[0] {
							if k != "_raw" {
								otherfield = true
								break
							}
						}
					}
					if otherfield {
						s.Tokens[i].FieldChoice = c.Samples[j].Lines
					} else {
						// s.Tokens[i].WeightedChoice = c.Samples[j].Lines
						temp := make([]string, 0, len(c.Samples[j].Lines))
						for _, line := range c.Samples[j].Lines {
							if _, ok := line["_raw"]; ok {
								if len(line["_raw"]) > 0 {
									temp = append(temp, line["_raw"])
								}
							}
						}
						s.Tokens[i].Choice = temp
					}
					break
				}
			}
		}

		// Begin Validation logic
		if s.EarliestParsed > s.LatestParsed {
			s.Log.Errorf("Earliest time cannot be greater than latest for sample '%s', disabling Sample", s.Name)
			s.Disabled = true
			return
		}
		// If no interval is set, generate one time and exit
		if s.Interval == 0 {
			s.Log.Infof("No interval set for sample '%s', setting endIntervals to 1", s.Name)
			s.EndIntervals = 1
		}
		for _, t := range s.Tokens {
			switch t.Type {
			case "random", "rated":
				if t.Replacement == "int" || t.Replacement == "float" {
					if t.Lower > t.Upper {
						s.Log.Errorf("Lower cannot be greater than Upper for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					} else if t.Upper == 0 {
						s.Log.Errorf("Upper cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else if t.Replacement == "string" || t.Replacement == "hex" {
					if t.Length == 0 {
						s.Log.Errorf("Length cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else {
					if t.Replacement != "guid" && t.Replacement != "ipv4" && t.Replacement != "ipv6" {
						s.Log.Errorf("Replacement '%s' is invalid for token '%s' in sample '%s'", t.Replacement, t.Name, s.Name)
						s.Disabled = true
					}
				}
			case "choice":
				if len(t.Choice) == 0 || t.Choice == nil {
					s.Log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "weightedChoice":
				if len(t.WeightedChoice) == 0 || t.WeightedChoice == nil {
					s.Log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "fieldChoice":
				if len(t.FieldChoice) == 0 || t.FieldChoice == nil {
					s.Log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
				for _, choice := range t.FieldChoice {
					if _, ok := choice[t.SrcField]; !ok {
						s.Log.Errorf("Source field '%s' does not exist for token '%s' in row '%#v' in sample '%s', disabling Sample", t.SrcField, t.Name, choice, s.Name)
						s.Disabled = true
						break
					}
				}
			}
		}
	}
}

func (c *Config) walkPath(fullPath string, acceptableExtensions map[string]bool, callback func(string) error) error {
	filepath.Walk(os.ExpandEnv(fullPath), func(path string, _ os.FileInfo, err error) error {
		c.Log.Debugf("Walking, at %s", path)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
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
			c.Log.Panicf("YAML parsing error in file '%s': %v", fullPath, err)
		}
	case ".json":
		if err := json.Unmarshal(contents, out); err != nil {
			c.Log.Panicf("JSON parsing errorin file '%s': %v", fullPath, err)
		}
	}
	// c.Log.Debugf("Out: %#v\n", out)
	return nil
}

func (c *Config) parseWebConfig(out interface{}, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Try YAML then JSON
	err = yaml.Unmarshal(contents, out)
	if err != nil {
		err = json.Unmarshal(contents, out)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindSampleByName finds and returns a pointer to a sample referenced by the passed name
func (c Config) FindSampleByName(name string) *Sample {
	for i := 0; i < len(c.Samples); i++ {
		if c.Samples[i].Name == name {
			return c.Samples[i]
		}
	}
	return nil
}
