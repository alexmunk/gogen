package outputter

import (
	"io"

	config "github.com/coccyx/gogen/internal"
)

type buf struct{}

func (foo buf) Send(item *config.OutQueueItem) error {
	_, err := io.Copy(item.S.Buf, item.IO.R)
	return err
}

func (foo buf) Close() error {
	return nil
}
