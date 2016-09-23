package config

// OutQueueItem represents one batch of events to output
type OutQueueItem struct {
	S      *Sample
	Events []map[string]string
}

// Outputter will output events using the designated output plugin
type Outputter interface {
	Send(item *OutQueueItem) error
}
