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
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
	"github.com/coccyx/timeparser"
	"github.com/ghodss/yaml"
	lua "github.com/yuin/gopher-lua"
)

// Config is a struct representing a Singleton which contains a copy of the running config
// across all processes.  Should mirror the structure of $GOGEN_HOME/configs/default/global.yml
type Config struct {
	Global      Global         `json:"global"`
	Samples     []*Sample      `json:"samples"`
	Templates   []*Template    `json:"templates"`
	Raters      []*RaterConfig `json:"raters"`
	initialized bool

	// Exported but internal use variables
	Timezone *time.Location `json:"-"`
	Buf      bytes.Buffer   `json:"-"`
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
	if instance == nil {
		instance = &Config{initialized: false}
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
		c = &Config{initialized: false}
		log.Debugf("Always refresh on, using fresh config")
	}

	// Setup timezone
	c.Timezone, _ = time.LoadLocation("Local")

	home := os.Getenv("GOGEN_HOME")
	if len(home) == 0 {
		log.Debug("GOGEN_HOME not set, setting to '.'")
		home = "."
		os.Setenv("GOGEN_HOME", home)
	}
	log.Debugf("Home: %v\n", home)

	samplesDir := os.Getenv("GOGEN_SAMPLES_DIR")
	if len(samplesDir) == 0 {
		samplesDir = filepath.Join(home, "config", "samples")
		log.Debugf("GOGEN_SAMPLES_DIR not set, setting to '%s'", samplesDir)
	}

	fullConfig := os.Getenv("GOGEN_FULLCONFIG")
	if len(fullConfig) > 0 {
		if fullConfig[0:4] == "http" {
			log.Infof("Fetching config from '%s'", fullConfig)
			if err := c.parseWebConfig(&c, fullConfig); err != nil {
				log.Panic(err)
			}
		} else {
			_, err := os.Stat(fullConfig)
			if err != nil {
				log.Fatalf("Cannot stat file %s", fullConfig)
			}
			if err := c.parseFileConfig(&c, fullConfig); err != nil {
				log.Panic(err)
			}
			// This seems like it might cause a regression, just commenting for now instead of removing
			// if filepath.Dir(fullConfig) != "." {
			// 	c.Global.SamplesDir = append(c.Global.SamplesDir, filepath.Dir(fullConfig))
			// }
		}
		for i := 0; i < len(c.Samples); i++ {
			c.Samples[i].realSample = true
		}
	} else {
		globalFile := os.Getenv("GOGEN_GLOBAL")
		if len(globalFile) > 0 {
			if err := c.parseFileConfig(&c.Global, globalFile); err != nil {
				log.Panic(err)
			}
		}
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
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			_ = template.New(t.Name+"_header", t.Row)
			_ = template.New(t.Name+"_row", t.Row)
			_ = template.New(t.Name+"_footer", t.Footer)

			c.Templates = append(c.Templates, t)
			return nil
		})

		// Read all raters in $GOGEN_HOME/config/raters
		fullPath = filepath.Join(home, "config", "raters")
		acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
		c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
			var r RaterConfig

			if err := c.parseFileConfig(&r, innerPath); err != nil {
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			c.Raters = append(c.Raters, &r)
			return nil
		})

		c.readSamplesDir(samplesDir)
	}

	// Configuration allows for finding additional samples directories and reading them
	for _, sd := range c.Global.SamplesDir {
		log.Debugf("Reading samplesDir from Global SamplesDir: %s", sd)
		c.readSamplesDir(sd)
	}

	// Add a clause to allow copying from other samples
	for i := 0; i < len(c.Samples); i++ {
		if len(c.Samples[i].FromSample) > 0 {
			for j := 0; j < len(c.Samples); j++ {
				if c.Samples[j].Name == c.Samples[i].FromSample {
					log.Debugf("Copying sample '%s' to sample '%s' because fromSample set", c.Samples[j].Name, c.Samples[i].Name)
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

	// Raters brought in from config will be typed wrong, validate and fixes
	for i := 0; i < len(c.Raters); i++ {
		c.validateRater(c.Raters[i])
	}

	// Due to data structure differences, we append default raters later in the startup process
	if os.Getenv("GOGEN_EXPORT") != "1" {
		raters := []*RaterConfig{defaultRaterConfig, defaultConfigRaterConfig}
		for _, r := range raters {
			c.Raters = append(c.Raters, r)
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

func (c *Config) readSamplesDir(samplesDir string) {
	// Read all flat file samples
	acceptableExtensions := map[string]bool{".sample": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
		s := new(Sample)
		s.Name = filepath.Base(innerPath)
		s.Disabled = true

		file, err := os.Open(innerPath)
		if err != nil {
			log.Errorf("Error reading sample file '%s': %s", innerPath, err)
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
			log.Errorf("Error reading sample file '%s': %s", innerPath, err)
			return nil
		}
		defer file.Close()

		reader := csv.NewReader(file)
		if fields, err = reader.Read(); err != nil {
			log.Errorf("Error parsing header row of sample file '%s' as csv: %s", innerPath, err)
			return nil
		}
		if rows, err = reader.ReadAll(); err != nil {
			log.Errorf("Error parsing sample file '%s' as csv: %s", innerPath, err)
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
			log.Errorf("Error parsing config %s: %s", innerPath, err)
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
		} else if len(s.Lines) == 0 {
			s.Disabled = true
			s.realSample = false
			log.Errorf("Disabling sample '%s', no lines in sample", s.Name)
		} else {
			s.realSample = true
		}

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
		if s.RaterString == "" {
			s.RaterString = defaultRater
		}

		ParseBeginEnd(s)

		//
		// Parse earliest and latest as relative times
		//
		// Cache a time so we can get a delta for parsed begin, end, earliest and latest
		n := time.Now()
		now := func() time.Time {
			return n
		}
		if p, err := timeparser.TimeParserNow(s.Earliest, now); err != nil {
			log.Errorf("Error parsing earliest time '%s' for sample '%s', using Now", s.Earliest, s.Name)
			s.EarliestParsed = time.Duration(0)
		} else {
			s.EarliestParsed = n.Sub(p) * -1
		}
		if p, err := timeparser.TimeParserNow(s.Latest, now); err != nil {
			log.Errorf("Error parsing latest time '%s' for sample '%s', using Now", s.Latest, s.Name)
			s.LatestParsed = time.Duration(0)
		} else {
			s.LatestParsed = n.Sub(p) * -1
		}

		// log.Debugf("Resolving '%s'", s.Name)
		for i := 0; i < len(s.Tokens); i++ {
			if s.Tokens[i].Type == "rated" && s.Tokens[i].RaterString == "" {
				s.Tokens[i].RaterString = "default"
			}
			if s.Tokens[i].Field == "" {
				s.Tokens[i].Field = s.Field
			}
			// If format is template, then create a default token of $tokenname$
			if s.Tokens[i].Format == "template" && s.Tokens[i].Token == "" {
				s.Tokens[i].Token = "$" + s.Tokens[i].Name + "$"
			}
			// log.Debugf("Resolving token '%s' for sample '%s'", s.Tokens[i].Name, s.Name)
			for j := 0; j < len(c.Samples); j++ {
				s.Tokens[i].Parent = s
				s.Tokens[i].luaState = new(lua.LTable)
				if s.Tokens[i].SampleString == c.Samples[j].Name {
					log.Debugf("Resolving sample '%s' for token '%s'", c.Samples[j].Name, s.Tokens[i].Name)
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

		if os.Getenv("GOGEN_EXPORT") != "1" && c.Global.Output.OutputTemplate == "splunkhec" {
			// If there's no _time token, add it to make sure we have a timestamp field in every event
			// This is primarily used for Splunk's HTTP Event Collectot
			timetoken := false
			for _, t := range s.Tokens {
				if t.Name == "_time" {
					timetoken = true
				}
			}
			for _, l := range s.Lines {
				if _, ok := l["_time"]; ok {
					timetoken = true
				}
			}
			if !timetoken {
				for i := 0; i < len(s.Lines); i++ {
					s.Lines[i]["_time"] = "$_time$"
				}
				tt := Token{
					Name:   "_time",
					Type:   "epochtimestamp",
					Format: "template",
					Field:  "_time",
					Token:  "$_time$",
					Group:  -1,
				}
				s.Tokens = append(s.Tokens, tt)
			}
			// Fixup existing timestamp tokens to all use the same static group, 999999
			for i := 0; i < len(s.Tokens); i++ {
				if s.Tokens[i].Type == "timestamp" || s.Tokens[i].Type == "gotimestamp" || s.Tokens[i].Type == "epochtimestamp" {
					s.Tokens[i].Group = -1
				}
			}
		}

		// Begin Validation logic
		if s.EarliestParsed > s.LatestParsed {
			log.Errorf("Earliest time cannot be greater than latest for sample '%s', disabling Sample", s.Name)
			s.Disabled = true
			return
		}
		// If no interval is set, generate one time and exit
		if s.Interval == 0 {
			log.Infof("No interval set for sample '%s', setting endIntervals to 1", s.Name)
			s.EndIntervals = 1
		}
		for _, t := range s.Tokens {
			switch t.Type {
			case "random", "rated":
				if t.Replacement == "int" || t.Replacement == "float" {
					if t.Lower > t.Upper {
						log.Errorf("Lower cannot be greater than Upper for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					} else if t.Upper == 0 {
						log.Errorf("Upper cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else if t.Replacement == "string" || t.Replacement == "hex" {
					if t.Length == 0 {
						log.Errorf("Length cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else {
					if t.Replacement != "guid" && t.Replacement != "ipv4" && t.Replacement != "ipv6" {
						log.Errorf("Replacement '%s' is invalid for token '%s' in sample '%s'", t.Replacement, t.Name, s.Name)
						s.Disabled = true
					}
				}
			case "choice":
				if len(t.Choice) == 0 || t.Choice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "weightedChoice":
				if len(t.WeightedChoice) == 0 || t.WeightedChoice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "fieldChoice":
				if len(t.FieldChoice) == 0 || t.FieldChoice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
				for _, choice := range t.FieldChoice {
					if _, ok := choice[t.SrcField]; !ok {
						log.Errorf("Source field '%s' does not exist for token '%s' in row '%#v' in sample '%s', disabling Sample", t.SrcField, t.Name, choice, s.Name)
						s.Disabled = true
						break
					}
				}
			}
		}

		// Check if we are able to do singlepass on this sample by looping through all lines
		// and ensuring we can match all the tokens on each line
		if !s.Disabled {
			s.SinglePass = true

			var tlines []map[string]tokenspos

		outer:
			for _, l := range s.Lines {
				tp := make(map[string]tokenspos)
				for j, t := range s.Tokens {
					// tokenpos 0 first char, 1 last char, 2 token #
					var pos tokenpos
					var err error
					pos1, pos2, err := t.GetReplacementOffsets(l[t.Field])
					if err != nil {
						log.Infof("Error getting replacements for token '%s' in event '%s', disabling SinglePass", t.Token, l[t.Field])
						s.SinglePass = false
						break outer
					}
					if pos1 < 0 || pos2 < 0 {
						log.Infof("Token '%s' not found in event '%s', disabling SinglePass", t.Name, l)
						s.SinglePass = false
						break outer
					}
					pos.Pos1 = pos1
					pos.Pos2 = pos2
					pos.Token = j
					tp[t.Field] = append(tp[t.Field], pos)
				}

				// Ensure we don't have any tokens overlapping one another for singlepass
				for _, v := range tp {
					sort.Sort(v)

					lastpos := 0
					lasttoken := ""
					maxpos := 0
					for _, pos := range v {
						// Does the beginning of this token overlap with the end of the last?
						if lastpos > pos.Pos1 {
							log.Infof("Token '%s' extends beyond beginning of token '%s', disabling SinglePass", lasttoken, s.Tokens[pos.Token].Name)
							s.SinglePass = false
							break outer
						}
						// Does the beginning of this token happen before the max we've seen a token before?
						if maxpos > pos.Pos1 {
							log.Infof("Some former token extends beyond the beginning of token '%s', disabling SinglePass", s.Tokens[pos.Token].Name)
							s.SinglePass = false
							break outer
						}
						if pos.Pos2 > maxpos {
							maxpos = pos.Pos2
						}
						lastpos = pos.Pos2
						lasttoken = s.Tokens[pos.Token].Name
					}
				}
				tlines = append(tlines, tp)
			}

			if s.SinglePass {

				// Now loop through each line and each field, breaking it up according to the positions of the tokens
				for i, line := range s.Lines {
					if len(tlines) >= i && len(tlines) > 0 {
						bline := make(map[string][]StringOrToken)
						for field := range line {
							var bfield []StringOrToken
							// Field doesn't exist because no tokens hit that field
							if _, ok := tlines[i][field]; !ok {
								bf := StringOrToken{T: nil, S: line[field]}
								bfield = append(bfield, bf)
							} else {
								lastpos := 0
								// Here, we need to iterate through all the tokens and add StringOrToken for each match
								// Make sure we check for a token a pos 0, we'll put a token first
								for _, tp := range tlines[i][field] {
									if tp.Pos1 == 0 {
										bf := StringOrToken{T: &s.Tokens[tp.Token], S: ""}
										bfield = append(bfield, bf)
										lastpos = tp.Pos2
									} else {
										// Add string from end of last token to the beginning of this one
										bf := StringOrToken{T: nil, S: s.Lines[i][field][lastpos:tp.Pos1]}
										bfield = append(bfield, bf)
										// Add this token
										bf = StringOrToken{T: &s.Tokens[tp.Token], S: ""}
										bfield = append(bfield, bf)
										lastpos = tp.Pos2
									}
								}
								// Add the last string if the last token didn't cover to the end of the string
								if lastpos < len(s.Lines[i][field]) {
									bf := StringOrToken{T: nil, S: s.Lines[i][field][lastpos:]}
									bfield = append(bfield, bf)
								}
							}
							bline[field] = bfield
						}
						s.BrokenLines = append(s.BrokenLines, bline)
					}
				}
			}
		}

		if s.Generator == "replay" {
			// For replay, loop through all events, attempt to find a timestamp in each row, store sleep times in a data structure
			s.ReplayOffsets = make([]time.Duration, len(s.Lines))
			var lastts time.Time
			var avgOffset time.Duration
		outer2:
			for i := 0; i < len(s.Lines); i++ {
			inner2:
				for _, t := range s.Tokens {
					if t.Type == "timestamp" || t.Type == "gotimestamp" || t.Type == "epochtimestamp" {
						pos1, pos2, err := t.GetReplacementOffsets(s.Lines[i][t.Field])
						if err != nil {
							log.WithFields(log.Fields{
								"token":  t.Name,
								"sample": s.Name,
								"err":    err,
							}).Errorf("Error getting timestamp offsets, disabling sample")
							s.Disabled = true
							break outer2
						}
						ts, err := t.ParseTimestamp(s.Lines[i][t.Field][pos1:pos2])
						if err != nil {
							log.WithFields(log.Fields{
								"token":  t.Name,
								"sample": s.Name,
								"err":    err,
								"event":  s.Lines[0][t.Field],
							}).Errorf("Error parsing timestamp, disabling sample")
							s.Disabled = true
							break outer2
						}
						if i == 0 {
							s.ReplayOffsets[0] = time.Duration(0)
						} else {
							s.ReplayOffsets[i] = lastts.Sub(ts) * -1
							avgOffset = (avgOffset + s.ReplayOffsets[i]) / 2
						}
						lastts = ts
						break inner2
					}
				}
				s.ReplayOffsets[0] = avgOffset
			}
		}
	}
}

// Returns a copy of the rater with the Options properly cast
func (c *Config) validateRater(r *RaterConfig) {
	configRaterKeys := map[string]bool{
		"HourOfDay":    true,
		"MinuteOfHour": true,
		"DayOfWeek":    true,
	}

	opt := make(map[string]interface{})
	for k, v := range r.Options {
		var newvset interface{}
		if configRaterKeys[k] {
			newv := make(map[int]float64)
			vcast := v.(map[string]interface{})
			for k2, v2 := range vcast {
				k2int, err := strconv.Atoi(k2)
				if err != nil {
					log.Fatalf("Rater key '%s' for rater '%s' in '%s' is not an integer value", k2, r.Name, k)
				}
				v2float, ok := v2.(float64)
				if !ok {
					log.Fatalf("Rater value '%#v' of key '%s' for rater '%s' in '%s' is not an integer value", v2, k2, r.Name, k)
				}
				newv[k2int] = v2float
			}
			newvset = newv
		} else {
			newvset = v
		}
		opt[k] = newvset
	}
	r.Options = opt
}

// FindRater returns a RaterConfig matched by the passed name
func (c *Config) FindRater(name string) *RaterConfig {
	for _, findr := range c.Raters {
		if findr.Name == name {
			return findr
		}
	}
	return nil
}

// ParseBeginEnd parses the Begin and End settings for a sample
func ParseBeginEnd(s *Sample) {
	// Setup Begin & End
	// If End is not set, then we're intended to always run in realtime
	if s.End == "" {
		s.Realtime = true
	}
	if s.Begin != "" && s.EndIntervals > 0 {
		s.Realtime = false
	}
	// Cache a time so we can get a delta for parsed begin, end, earliest and latest
	n := time.Now()
	now := func() time.Time {
		return n
	}
	var err error
	if len(s.Begin) > 0 {
		if s.BeginParsed, err = timeparser.TimeParserNow(s.Begin, now); err != nil {
			log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
		} else {
			s.Current = s.BeginParsed
			s.Realtime = false
		}
	}
	if len(s.End) > 0 {
		if s.EndParsed, err = timeparser.TimeParserNow(s.End, now); err != nil {
			log.Errorf("Error parsing End for sample %s: %v", s.Name, err)
		}
	}
	log.Infof("Beginning generation at %s; Ending at %s; Realtime: %v", s.BeginParsed, s.EndParsed, s.Realtime)
}

func (c *Config) walkPath(fullPath string, acceptableExtensions map[string]bool, callback func(string) error) error {
	filepath.Walk(os.ExpandEnv(fullPath), func(path string, _ os.FileInfo, err error) error {
		log.Debugf("Walking, at %s", path)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			log.Errorf("Error from WalkFunc: %s", err)
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
	log.Debugf("Config Path: %v\n", fullPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return err
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// log.Debugf("Contents: %s", contents)
	switch filepath.Ext(fullPath) {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(contents, out); err != nil {
			log.Panicf("YAML parsing error in file '%s': %v", fullPath, err)
		}
	case ".json":
		if err := json.Unmarshal(contents, out); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				log.Panicf("JSON parsing error in file '%s' at offset %d: %v", fullPath, ute.Offset, ute)
			} else {
				log.Panicf("JSON parsing error in file '%s': %v", fullPath, err)
			}
		}
	}
	// log.Debugf("Out: %#v\n", out)
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
