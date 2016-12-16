package internal

// Template is used for describing a output text template
type Template struct {
	Name   string `json:"name" yaml:"name"`
	Header string `json:"header,omitempty" yaml:"header,omitempty"`
	Row    string `json:"row" yaml:"row"`
	Footer string `json:"footer,omitempty" yaml:"footer,omitempty"`
}
