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

func TestLuaGen(t *testing.T) {
	config.ResetConfig()
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "realGenerator.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("generator")

	gen := new(luagen)
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 walked in")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 sat down")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 on the group w bench")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 walked in")
}

func testLuaGen(t *testing.T, s *config.Sample, gen *luagen, expected string) {
	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen}
	var err error
	go func() {
		err = gen.Gen(gqi)
	}()
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	var good bool
	good = false
	select {
	case oqi := <-oq:
		assert.Equal(t, expected, oqi.Events[0]["_raw"])
		good = true
	case <-timeout:
		if !good {
			t.Fatalf("Timed out, err: %s", err)
		}
	}
}
