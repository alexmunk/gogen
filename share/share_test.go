package share

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestSharePush(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "examples", "weblog", "weblog.yml"))
	_ = config.NewConfig()
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	os.Setenv("GOGEN_EXPORT", "")
	Push("weblog")
}

func TestSharePull(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	_ = os.Mkdir("testout", 0777)
	Pull("coccyx/weblog", "testout", false)
	_, err := os.Stat("testout/weblog.json")
	assert.NoError(t, err, "Couldn't find file weblog.json")
	_ = os.Remove("testout/weblog.json")

	Pull("coccyx/weblog", "testout", true)
	_, err = os.Stat("testout/samples/weblog.json")
	assert.NoError(t, err, "Couldn't find file samples/weblog.json")
	_, err = os.Stat("testout/samples/webhosts.csv")
	assert.NoError(t, err, "Couldn't find file samples/webhosts.csv")
	_, err = os.Stat("testout/samples/useragents.sample")
	assert.NoError(t, err, "Couldn't find file samples/useragents.sample")
	_ = os.RemoveAll("testout")

}
