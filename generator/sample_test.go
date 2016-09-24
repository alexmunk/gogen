package generator

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestSampleGen(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "tokens.yml"))
	loc, _ := time.LoadLocation("Local")
	rand.Seed(0)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	s := FindSampleInFile(home, "token-static")
	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), S: s, OQ: oq}
	gen := new(sample)
	go gen.Gen(gqi)
	oqi := <-oq
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])
}

func FindSampleInFile(home string, name string) *config.Sample {
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", name+".yml"))
	c := config.NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	return c.FindSampleByName(name)
}
