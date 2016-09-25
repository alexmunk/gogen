package outputter

import (
	"fmt"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/template"
)

type stdout struct{}

func (foo stdout) Send(item *config.OutQueueItem) error {
	// We'll crash on empty events, but don't do that!
	sendline("header", item.S, item.Events[0])
	// item.S.Log.Debugf("Out Queue Item %#v", item)
	var last int
	for i, line := range item.Events {
		sendline("row", item.S, line)
		last = i
	}
	sendline("footer", item.S, item.Events[last])
	return nil
}

func sendline(templatename string, s *config.Sample, line map[string]string) {
	if template.Exists(s.OutputTemplate + "_" + templatename) {
		linestr, err := template.Exec(s.OutputTemplate+"_"+templatename, line)
		if err != nil {
			s.Log.Errorf("Error from sample '%s' in template execution: %v", s.Name, err)
		} else {
			// item.S.Log.Debugf("Outputting line %s", linestr)
			fmt.Println(linestr)
		}
	}
}
