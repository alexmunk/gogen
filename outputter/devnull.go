package outputter

import (
	"io"
	"io/ioutil"

	config "github.com/coccyx/gogen/internal"
)

type devnull struct{}

func (foo devnull) Send(item *config.OutQueueItem) error {
	_, err := io.Copy(ioutil.Discard, item.IO.R)
	return err
}

func (foo devnull) Close() error {
	return nil
}
