package outputter

import (
	"net/http"

	config "github.com/coccyx/gogen/internal"
)

type httpout struct {
	initialized bool
	client      *http.Client
	req         *http.Request
}

func (h *httpout) Send(item *config.OutQueueItem) error {
	// if h.initialized == false {
	// 	h.client = &http.Client{}
	// 	h.req = http.NewRequest("POST")

	// 	h.initialized = true
	// }
	// bytes, err := io.Copy(f.file, item.IO.R)

	// Account(int64(len(item.Events)), bytes)
	// f.fileSize += bytes
	// if f.fileSize >= item.S.Output.MaxBytes {
	// 	item.S.Log.Infof("Reached %d bytes which exceeds MaxBytes for sample '%s', rotating files", f.fileSize, item.S.Name)
	// 	f.rotate(item)
	// }
	// return err
	return nil
}

// client := &http.Client{
// 	CheckRedirect: redirectPolicyFunc,
// }

// resp, err := client.Get("http://example.com")
// // ...

// req, err := http.NewRequest("GET", "http://example.com", nil)
// // ...
// req.Header.Add("If-None-Match", `W/"wyzzy"`)
// resp, err := client.Do(req)
