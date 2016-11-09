package generator

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestPrimeRater(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "luarater.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("double")
	PrimeRater(s)
	for _, token := range s.Tokens {
		if token.Name == "multiply" {
			assert.NotNil(t, token.Rater)
		}
	}
}
