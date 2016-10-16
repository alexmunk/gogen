package outputter

import (
	"io"

	config "github.com/coccyx/gogen/internal"
)

type buf struct{}

func (foo buf) Send(item *config.OutQueueItem) error {
	c := config.NewConfig()
	if c.Buf.Len() > 0 {
		_, _ = c.Buf.WriteString("\n")
	}
	_, err := io.Copy(&c.Buf, item.IO.R)
	return err
}

func (foo buf) Close() error {
	return nil
}
