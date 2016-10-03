package config

import logging "github.com/op/go-logging"

const defaultLoggingLevel = logging.ERROR

// Default global values
const defaultGeneratorWorkers = 1
const defaultOutputWorkers = 1
const defaultROTInterval = 1
const defaultOutputter = "stdout"
const defaultOutputTemplate = "raw"

// Default Sample values
const defaultGenerator = "sample"
const defaultEarliest = "now"
const defaultLatest = "now"
const defaultRandomizeEvents = true
const defaultField = "_raw"

// Default file output values
const defaultFileName = "/tmp/test.log"
const defaultMaxBytes = 10485760
const defaultBackupFiles = 5

// Default HTTP output values
const defaultBufferBytes = 102400

// MaxOutputThreads defines how large an array we'll define for output threads
const MaxOutputThreads = 100

// MaxGenQueueLength defines how many items can be in the Generator queue at a given time
const MaxGenQueueLength = 100

// MaxOutQueueLength defines how many items can be in the Output queue at a given time
const MaxOutQueueLength = 100

var (
	defaultCSVTemplate       *Template
	defaultJSONTemplate      *Template
	defaultSplunkHECTemplate *Template
	defaultRawTemplate       *Template
)

func init() {
	defaultCSVTemplate = &Template{
		Name:   "csv",
		Header: `{{ keys . | join "," }}`,
		Row:    `{{ values . | join "," }}`,
		Footer: "",
	}
	defaultJSONTemplate = &Template{
		Name:   "json",
		Header: "",
		Row:    `{{ json . | printf "%s" }}`,
		Footer: "",
	}
	defaultSplunkHECTemplate = &Template{
		Name:   "splunkhec",
		Header: "",
		Row:    `{{ splunkhec . | printf "%s" }}`,
		Footer: "",
	}
	defaultRawTemplate = &Template{
		Name:   "raw",
		Header: "",
		Row:    `{{ ._raw }}`,
		Footer: "",
	}
}
