package internal

// Mix is a list of configurations and overrides for those configurations to allow users to assemble new derived configurations from a mix of other existing configs
type Mix struct {
	Sample       string `json:"sample" yaml:"sample"`
	Interval     int    `json:"interval,omitempty" yaml:"interval,omitempty"`
	Count        int    `json:"count,omitempty" yaml:"count,omitempty"`
	Begin        string `json:"begin,omitempty" yaml:"begin,omitempty"`
	End          string `json:"end,omitempty" yaml:"end,omitempty"`
	EndIntervals int    `json:"endIntervals,omitempty" yaml:"endIntervals,omitempty"`
}
