package outputter

import (
	"io"
	"os"

	config "github.com/coccyx/gogen/internal"
)

type stdout struct{}

func (foo stdout) Send(item *config.OutQueueItem) error {
	_, err := io.Copy(os.Stdout, item.IO.R)

	return err
}

func (foo stdout) Close() error {
	return nil
}
