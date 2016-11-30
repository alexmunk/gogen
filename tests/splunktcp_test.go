package tests

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
)

func TestSplunkTCPOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "splunktcpoutput", "splunktcpoutput.yml"))
	c := config.NewConfig()
	s := c.FindSampleByName("outputsample")
	if s != nil {
		run.Run(c)
	}
	run.Run(c)
}
