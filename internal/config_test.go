package config

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	// Test Singleton
	c := NewConfig()
	c2 := NewConfig()
	assert.Equal(t, c, c2)

	global := Global{Debug: false, Verbose: false, UseOutputQueue: true, GeneratorWorkers: 1, OutputWorkers: 1}
	assert.Equal(t, c.Global, global)
	defaultSample := Sample{Name: "", Disabled: false, Generator: "sample", Outputter: "stdout", OutputTemplate: "raw", Rater: "config", Interval: 60, Delay: 0, Count: 0, Earliest: "now", Latest: "now", Begin: "", End: "", RandomizeCount: 0.2, RandomizeEvents: true, Field: "_raw"}
	assert.Equal(t, c.DefaultSample, defaultSample)
}

func TestValidate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	rand.Seed(0)
	// loc, _ := time.LoadLocation("Local")
	// n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	// now := func() time.Time {
	// 	return n
	// }

	var s *Sample
	s = FindSampleInFile(home, "validate-lower-upper")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-upper")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-string-length")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-choice-items")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-weightedchoice-items")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-fieldchoice-items")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-fieldchoice-badfield")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-badrandom")
	assert.Equal(t, true, s.Disabled)

	s = FindSampleInFile(home, "validate-earliest-latest")
	assert.Equal(t, true, s.Disabled)
}

func FindSampleInFile(home string, name string) *Sample {
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", name+".yml"))
	c := NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	return c.FindSampleByName(name)
}
