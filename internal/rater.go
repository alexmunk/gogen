package config

import "time"

// RaterConfig defines how to rate an event or token
type RaterConfig struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Script  string                 `json:"script,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
	Init    map[string]string      `json:"init,omitempty"`
}

// Rater will rate an event according to RaterConfig
type Rater interface {
	GetRate(now time.Time) float64
	EventRate(s *Sample, now time.Time, count int) int
	TokenRate(t Token, now time.Time) float64
}
