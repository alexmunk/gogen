package outputter

import (
	"io"
	"io/ioutil"

	config "github.com/coccyx/gogen/internal"
)

type devnull struct{}

func (foo devnull) Send(item *config.OutQueueItem) error {
	bytes, err := io.Copy(ioutil.Discard, item.IO.R)

	Account(int64(len(item.Events)), bytes)
	return err
}
