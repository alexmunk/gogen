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
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	s := config.FindSampleInFile(home, "token-static")
	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), S: s, OQ: oq, Rand: randgen}
	gen := new(sample)
	go gen.Gen(gqi)
	oqi := <-oq
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])

	s = config.FindSampleInFile(home, "token-regex")
	gqi = &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), S: s, OQ: oq, Rand: randgen}
	gen = new(sample)
	go gen.Gen(gqi)
	oqi = <-oq
	assert.Equal(t, "foo foo bar", oqi.Events[0]["_raw"])
}
