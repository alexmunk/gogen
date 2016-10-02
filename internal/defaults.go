package config

import logging "github.com/op/go-logging"

const defaultLoggingLevel = logging.DEBUG

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

var (
	defaultCSVTemplate  *Template
	defaultJSONTemplate *Template
	defaultRawTemplate  *Template
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
	defaultRawTemplate = &Template{
		Name:   "raw",
		Header: "",
		Row:    `{{ ._raw }}`,
		Footer: "",
	}
}
