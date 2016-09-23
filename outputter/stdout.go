package outputter

import (
	"fmt"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/template"
)

type stdout struct{}

func (foo stdout) Send(item *config.OutQueueItem) error {
	item.S.Log.Debugf("Out Queue Item %#v", item)
	for _, line := range item.Events {
		linestr, err := template.Exec(item.S.OutputTemplate+"_row", line)
		if err != nil {
			item.S.Log.Errorf("Error from sample '%s' in template execution: %v", item.S.Name, err)
		} else {
			fmt.Println(linestr)
		}
	}
	return nil
}
