package generator

import (
	"fmt"

	"github.com/coccyx/gogen/config"
)

type sample struct{}

func (foo sample) Gen(item *config.GenQueueItem) error {
	item.S.Log.Debugf("Gen Queue Item %#v", item)
	outstr := []map[string]string{{"_raw": fmt.Sprintf("%#v", item)}}
	outitem := &config.OutQueueItem{S: item.S, Events: outstr}
	select {
	case item.OQ <- outitem:
	default:
	}
	return nil
}
