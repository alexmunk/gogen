package outputter

import "github.com/coccyx/gogen/config"

type stdout struct{}

func (foo stdout) Send(item *config.OutQueueItem) error {
	item.S.Log.Debugf("Out Queue Item %#v", item)
	return nil
}
