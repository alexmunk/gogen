package tests

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestFileOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_GLOBAL", filepath.Join(home, "config", "tests", "fileoutput.yml"))
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "outputsample.yml"))
	c := config.NewConfig()
	// s := c.FindSampleByName("backfill")
	run.Run(c)

	info, err := os.Stat(c.Global.Output.FileName)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		if info.Size() < c.Global.Output.MaxBytes {
			return true
		}
		return false
	}, "Rotation failing, main file size of %d greater than MaxBytes %d", info.Size(), c.Global.Output.MaxBytes)
	for i := 1; i <= c.Global.Output.BackupFiles; i++ {
		info, err = os.Stat(c.Global.Output.FileName + "." + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Condition(t, func() bool {
			if info.Size() > c.Global.Output.MaxBytes {
				return true
			}
			return false
		}, "Rotation failing, file %d less than MaxBytes", i)
	}

}
