package share

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
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
