package timer

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	c := config.NewConfig()
	s := c.FindSampleByName("translog")
	gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)

	timer := &Timer{S: s, GQ: gq, OQ: oq}
	go timer.NewTimer()

	item := <-gq

	// Test that we get a GenQueueItem
	var gqi *config.GenQueueItem
	assert.Equal(t, reflect.TypeOf(gqi), reflect.ValueOf(item).Type())

	// Test that we're about the same interval
	n := time.Now()
	timer = &Timer{S: s, GQ: gq, OQ: oq}
	go timer.NewTimer()
	item = <-gq
	cur := time.Now()

	gt := cur.Sub(n) > (time.Duration(s.Interval) * time.Second)
	assert.Equal(t, true, gt)
}
