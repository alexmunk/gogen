package tests

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
)

func TestHTTPOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "httpoutput", "httpoutput.yml"))
	c := config.NewConfig()
	// s := c.FindSampleByName("backfill")
	run.Run(c)
	// open.Run(c.Global.Output.Endpoints[0] + "?inspect")

	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "httpoutput", "splunkoutput.yml"))
	c = config.NewConfig()
	run.Run(c)
}
