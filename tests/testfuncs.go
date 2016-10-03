package tests

import (
	"os"
	"path/filepath"

	config "github.com/coccyx/gogen/internal"
)

// FindSampleInFile is used for tests to find a specific test file and reload config with it
func FindSampleInFile(home string, name string) *config.Sample {
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, name+".yml"))
	c := config.NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	return c.FindSampleByName(name)
}
