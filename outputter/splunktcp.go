package outputter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"encoding/binary"

	config "github.com/coccyx/gogen/internal"
)

type splunktcp struct {
	buf         *bufio.Writer
	r           *io.PipeReader
	w           *io.PipeWriter
	conn        net.Conn
	initialized bool
	endpoint    string
	closed      bool
	lastS       *config.Sample
	sent        int64
	done        chan int
}

type splunkSignature struct {
	signature  [128]byte
	serverName [256]byte
	mgmtPort   [16]byte
}

var bp sync.Pool

func init() {
	bp = sync.Pool{
		New: func() interface{} {
			bb := bytes.NewBuffer([]byte{})
			return bb
		},
	}
}

// connect opens a connection to Splunk
func (st *splunktcp) connect(endpoint string) error {
	var err error
	st.conn, err = net.DialTimeout("tcp", endpoint, 2*time.Second)
	return err
}

// sendSig will write the signature to the connection if it has not already been written
// Create Signature element of the S2S Message.  Signature is C struct:
//
// struct S2S_Signature
// {
// 	char _signature[128];
// 	char _serverName[256];
// 	char _mgmtPort[16];
// };
func (st *splunktcp) sendSig() error {
	endpointParts := strings.Split(st.endpoint, ":")
	if len(endpointParts) != 2 {
		return fmt.Errorf("Endpoint malformed.  Should look like server:port")
	}
	serverName := endpointParts[0]
	mgmtPort := endpointParts[1]
	var sig splunkSignature
	copy(sig.signature[:], "--splunk-cooked-mode-v2--")
	copy(sig.serverName[:], serverName)
	copy(sig.mgmtPort[:], mgmtPort)
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, sig.signature)
	binary.Write(buf, binary.BigEndian, sig.serverName)
	binary.Write(buf, binary.BigEndian, sig.mgmtPort)
	st.buf.Write(buf.Bytes())
	return nil
}

// encodeString encodes a string to be sent across the wire to Splunk
// Wire protocol has an unsigned integer of the length of the string followed
// by a null terminated string.
func encodeString(tosend string) []byte {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf := &bytes.Buffer{}
	l := uint32(len(tosend) + 1)
	binary.Write(buf, binary.BigEndian, l)
	binary.Write(buf, binary.BigEndian, []byte(tosend))
	binary.Write(buf, binary.BigEndian, []byte{0})
	return buf.Bytes()
}

// encodeKeyValue encodes a key/value pair to send across the wire to splunk
// A key value pair is merely a concatenated set of encoded strings.
func encodeKeyValue(key, value string) []byte {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf := &bytes.Buffer{}
	buf.Write(encodeString(key))
	buf.Write(encodeString(value))
	return buf.Bytes()
}

// encodeEvent encodes a full Splunk event
func encodeEvent(line map[string]string) []byte {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf := &bytes.Buffer{}

	var msgSize uint32
	msgSize = 8 // Two unsigned 32 bit integers included, the number of maps and a 0 between the end of raw the _raw trailer
	maps := make([][]byte, 0)

	for k, v := range line {
		switch k {
		case "source":
			encodedSource := encodeKeyValue("MetaData:Source", "source::"+v)
			maps = append(maps, encodedSource)
			msgSize += uint32(len(encodedSource))
		case "sourcetype":
			encodedSourcetype := encodeKeyValue("MetaData:Sourcetype", "sourcetype::"+v)
			maps = append(maps, encodedSourcetype)
			msgSize += uint32(len(encodedSourcetype))
		case "host":
			encodedHost := encodeKeyValue("MetaData:Host", "host::"+v)
			maps = append(maps, encodedHost)
			msgSize += uint32(len(encodedHost))
		case "index":
			encodedIndex := encodeKeyValue("_MetaData:Index", v)
			maps = append(maps, encodedIndex)
			msgSize += uint32(len(encodedIndex))
		case "_raw":
			break
		default:
			encoded := encodeKeyValue(k, v)
			maps = append(maps, encoded)
			msgSize += uint32(len(encoded))
		}
	}

	encodedRaw := encodeKeyValue("_raw", line["_raw"])
	msgSize += uint32(len(encodedRaw))
	encodedRawTrailer := encodeString("_raw")
	msgSize += uint32(len(encodedRawTrailer))
	encodedDone := encodeKeyValue("_done", "_done")
	msgSize += uint32(len(encodedDone))

	binary.Write(buf, binary.BigEndian, msgSize)
	binary.Write(buf, binary.BigEndian, uint32(len(maps)+2)) // Include extra map for _done key and one for _raw
	for _, m := range maps {
		binary.Write(buf, binary.BigEndian, m)
	}
	binary.Write(buf, binary.BigEndian, encodedDone)
	binary.Write(buf, binary.BigEndian, encodedRaw)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, encodedRawTrailer)

	return buf.Bytes()
}

func (st *splunktcp) Send(item *config.OutQueueItem) error {
	if st.initialized == false {
		err := st.newBuf(item)
		if err != nil {
			return err
		}
		err = st.sendSig()
		if err != nil {
			return err
		}
		st.initialized = true
	}
	bytes, err := io.Copy(st.buf, item.IO.R)
	if err != nil {
		return err
	}

	st.sent += bytes
	if st.sent > int64(item.S.Output.BufferBytes) {
		err := st.buf.Flush()
		if err != nil {
			return err
		}
		st.newBuf(item)
		st.sent = 0
	}
	st.lastS = item.S
	return nil
}

func (st *splunktcp) Close() error {
	if !st.closed {
		err := st.buf.Flush()
		if err != nil {
			return err
		}
		st.closed = true
	}
	return nil
}

func (st *splunktcp) newBuf(item *config.OutQueueItem) error {
	st.endpoint = item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
	err := st.connect(st.endpoint)
	if err != nil {
		return err
	}
	st.buf = bufio.NewWriter(st.conn)
	return nil
}
