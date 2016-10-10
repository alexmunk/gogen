package share

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

var (
	gh *GitHub
	id string
	c  *config.Config
	tc *config.Config
)

func TestLogin(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "examples", "weblog", "weblog.yml"))
	os.Setenv("GOGEN_EXPORT", "1")
	c = config.NewConfig()
	gh = NewGitHub(true)
	assert.NotNil(t, gh, "NewGitHub() returned nil")
}

func TestPush(t *testing.T) {
	_ = gh.Push("test_config")

	l, _, _ := gh.client.Gists.List("", nil)
	inList := false
	for _, item := range l {
		if *item.Description == "test_config" {
			inList = true
			id = *item.ID
		}
	}
	if !inList {
		t.Fatal("test_config not in Gist list")
	}
}

func TestValid(t *testing.T) {
	g, _, err := gh.client.Gists.Get(id)
	assert.NoError(t, err, "Failed getting gist")
	// fmt.Printf("%# v", pretty.Formatter(g))
	_, err = gh.client.Gists.Delete(id)
	assert.NoError(t, err, "Failed deleting gist")
	content := []byte(*g.Files["test_config.json"].Content)
	err = ioutil.WriteFile("test_config.json", content, 444)
	assert.NoError(t, err, "Cannot write test_config.json")

	os.Setenv("GOGEN_FULLCONFIG", "test_config.json")
	tc = config.NewConfig()
	_ = os.Remove("test_config.json")

	assert.Equal(t, c.Samples[0].Name, tc.Samples[0].Name)
}
