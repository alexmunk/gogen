package outputter

import (
	"bufio"
	"crypto/tls"
	"io"
	"math/rand"
	"net/http"

	config "github.com/coccyx/gogen/internal"
)

type httpout struct {
	buf         *bufio.Writer
	r           *io.PipeReader
	w           *io.PipeWriter
	resp        *http.Response
	initialized bool
	closed      bool
	lastS       *config.Sample
	sent        int64
	done        chan int
}

func (h *httpout) Send(item *config.OutQueueItem) error {
	if h.initialized == false {
		h.newPost(item)
		h.initialized = true
	}
	bytes, err := io.Copy(h.buf, item.IO.R)
	if err != nil {
		return err
	}

	h.sent += bytes
	if h.sent > int64(item.S.Output.BufferBytes) {
		err := h.buf.Flush()
		if err != nil {
			return err
		}
		err = h.w.Close()
		if err != nil {
			return err
		}
		h.newPost(item)
		h.sent = 0
	}
	h.lastS = item.S
	return nil
}

func (h *httpout) Close() error {
	if !h.closed {
		err := h.buf.Flush()
		if err != nil {
			return err
		}
		err = h.w.Close()
		if err != nil {
			return err
		}
		<-h.done
		h.closed = true
	}
	return nil
}

func (h *httpout) newPost(item *config.OutQueueItem) {
	h.r, h.w = io.Pipe()
	h.buf = bufio.NewWriter(h.w)

	endpoint := item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", endpoint, h.r)
	for k, v := range item.S.Output.Headers {
		req.Header.Add(k, v)
	}
	h.done = make(chan int)
	go func() {
		h.resp, err = client.Do(req)
		h.done <- 1
		if err != nil {
			item.S.Log.Errorf("Error making request from sample '%s' to endpoint '%s': %s", item.S.Name, endpoint, err)
		}
	}()
}
