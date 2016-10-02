package config

import (
	"os"
	"path/filepath"
)

// FindSampleInFile is used for tests to find a specific test file and reload config with it
func FindSampleInFile(home string, name string) *Sample {
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, name+".yml"))
	c := NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	return c.FindSampleByName(name)
}
