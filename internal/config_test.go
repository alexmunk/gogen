package config

import (
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleton(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	c := NewConfig()
	c2 := NewConfig()
	assert.Equal(t, c, c2)
}

func TestGlobal(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	c := NewConfig()
	output := Output{Outputter: "stdout", OutputTemplate: "raw"}
	global := Global{Debug: false, Verbose: false, GeneratorWorkers: 1, OutputWorkers: 1, ROTInterval: 1, Output: output}
	assert.Equal(t, global, c.Global)
}

func TestFileOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "fileoutput", "fileoutput.yml"))
	// os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "fileoutput.yml"))
	c := NewConfig()

	// Test flatten
	assert.Equal(t, "/tmp/fileoutput.log", c.Global.Output.FileName)
	assert.Equal(t, int64(102400), c.Global.Output.MaxBytes)
	assert.Equal(t, "file", c.Global.Output.Outputter)
	assert.Equal(t, "json", c.Global.Output.OutputTemplate)
}

func TestHTTPOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "httpoutput", "httpoutput.yml"))
	// os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "fileoutput.yml"))
	c := NewConfig()

	headers := map[string]string{"Authorization": "Splunk 00112233-4455-6677-8899-AABBCCDDEEFF"}
	endpoints := []string{"http://requestb.in/18yarph1"}
	de := reflect.DeepEqual(headers, c.Global.Output.Headers)
	assert.True(t, de, "Headers do not match: %#v vs %#v", headers, c.Global.Output.Headers)
	de = reflect.DeepEqual(endpoints, c.Global.Output.Endpoints)
	assert.True(t, de, "Endpoints do not match: %#v vs %#v", endpoints, c.Global.Output.Endpoints)
}

func TestFlatten(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "flatten")
	// os.Setenv("GOGEN_GLOBAL", filepath.Join(home, "config", "global.yml"))
	rand.Seed(0)

	var s *Sample
	s = FindSampleInFile(home, "flatten")
	assert.Equal(t, false, s.Disabled)
	assert.Equal(t, "sample", s.Generator)
	assert.Equal(t, "stdout", s.Output.Outputter)
	assert.Equal(t, "raw", s.Output.OutputTemplate)
	// assert.Equal(t, "config", s.Rater)
	assert.Equal(t, 0, s.Interval)
	assert.Equal(t, 0, s.Count)
	assert.Equal(t, "now", s.Earliest)
	assert.Equal(t, "now", s.Latest)
	// if diff := math.Abs(float64(0.2 - s.RandomizeCount)); diff > 0.000001 {
	// 	t.Fatalf("RandomizeCount not equal")
	// }
	assert.Equal(t, true, s.RandomizeEvents)
	assert.Equal(t, 1, s.EndIntervals)
}

func TestValidate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "validation")
	rand.Seed(0)
	// loc, _ := time.LoadLocation("Local")
	// n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	// now := func() time.Time {
	// 	return n
	// }

	var s *Sample
	checks := []string{
		"validate-lower-upper",
		"validate-upper",
		"validate-string-length",
		"validate-choice-items",
		"validate-weightedchoice-items",
		"validate-fieldchoice-items",
		"validate-fieldchoice-items",
		"validate-fieldchoice-badfield",
		"validate-badrandom",
		"validate-earliest-latest",
	}
	for _, v := range checks {
		s = FindSampleInFile(home, v)
		assert.Nil(t, s, "%s not nil", v)
	}
}
