package outputter

import (
	"io"
	"os"

	config "github.com/coccyx/gogen/internal"
)

type stdout struct{}

func (foo stdout) Send(item *config.OutQueueItem) error {
	bytes, err := io.Copy(os.Stdout, item.IO.R)

	Account(int64(len(item.Events)), bytes)
	return err
}

func (foo stdout) Close() error {
	return nil
}
