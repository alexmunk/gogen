package config

// Template is used for describing a output text template
type Template struct {
	Name   string `json:"name"`
	Header string `json:"header"`
	Row    string `json:"row"`
	Footer string `json:"footer"`
}
