package config

import (
	"io"
	"math/rand"
)

// OutQueueItem represents one batch of events to output
type OutQueueItem struct {
	S      *Sample
	Events []map[string]string
	Rand   *rand.Rand
	IO     *OutputIO
}

// OutputIO contains our Readers and Writers
type OutputIO struct {
	R io.Reader
	W io.WriteCloser
}

// NewOutputIO returns a freshly initialized pipe and TeeReader
func NewOutputIO() *OutputIO {
	o := new(OutputIO)
	o.R, o.W = io.Pipe()
	return o
}

// Outputter will do the work of actually sending events
type Outputter interface {
	Send(item *OutQueueItem) error
}
